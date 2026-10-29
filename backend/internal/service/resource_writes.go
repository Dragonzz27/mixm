package service

import (
	"context"
	"fmt"
	"net/http"

	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

// saveResource 校验资源及关联关系，并在事务中完成保存。
func (s *Core) saveResource(
	ctx context.Context,
	actor model.Actor,
	db *repository.DB,
	table, id string,
	v model.Record,
) (model.Record, error) {
	if err := validateResource(table, id, v); err != nil {
		return nil, err
	}

	if id == "" {
		if _, ok := repository.Columns[table]["created_by"]; ok {
			v["created_by"] = actor.ID
		}
	}

	delete(v, "reviewed_by")
	delete(v, "reviewed_at")
	if table == "questions" {
		v["review_status"] = "pending"
		v["reviewed_by"], v["reviewed_at"] = nil, nil
		delete(v, "is_ai_generated")
	}

	if table == "templates" {
		if _, ok := v["config"]; ok {
			if _, isString := v["config"].(string); !isString {
				v["config"] = model.JSON(v["config"])
			}
		}
	}

	if table == "assignments" {
		if _, ok := v["attachments"]; ok {
			v["attachments"] = model.JSON(v["attachments"])
		}
	}

	if table == "categories" || table == "question_categories" {
		if p, ok := v["parent_id"]; ok {
			if p == "" {
				p = nil
				v["parent_id"] = nil
			}

			v["level"] = 1
			if p != nil {
				if p == id {
					return nil, fail(http.StatusUnprocessableEntity, "分类不能作为自己的父级")
				}

				parent, err := repository.Get(ctx, db, table, fmt.Sprint(p))
				if err != nil {
					return nil, err
				}

				if id != "" {
					cycle, err := repository.CategoryCycle(ctx, db, table, fmt.Sprint(p), id)
					if err != nil {
						return nil, err
					}

					if model.Bool(cycle, "cycle") {
						return nil, fail(http.StatusUnprocessableEntity, "分类形成循环")
					}
				}

				v["level"] = int(model.Number(parent, "level")) + 1
			}
		}
	}

	out, err := repository.Write(ctx, db, table, id, v)
	if err != nil {
		return nil, err
	}

	id = model.String(out, "id")
	prefix := map[string]string{
		"questions":   "question",
		"practices":   "practice",
		"assignments": "assignment",
	}[table]

	for _, rel := range []struct{ field, suffix, col string }{
		{"visible_roles", "roles", "role_id"},
		{"knowledge_tag_ids", "tags", "category_id"},
	} {
		if vals, ok := v[rel.field]; ok && prefix != "" && !(table == "assignments" && rel.suffix == "tags") {
			rt := prefix + "_" + rel.suffix
			if _, err = repository.ClearResourceLinks(ctx, db, prefix, rt, id); err != nil {
				return nil, err
			}

			seen := map[string]bool{}

			for _, rid := range model.Strings(vals) {
				if seen[rid] {
					continue
				}

				seen[rid] = true
				if _, err = repository.InsertResourceLink(ctx, db, rel.col, prefix, rt, id, rid); err != nil {
					return nil, err
				}
			}
		}
	}

	if table == "questions" {
		for _, field := range []string{"options", "attachments"} {
			if vals, ok := v[field]; ok {
				rt := "question_" + field
				if _, err = repository.ClearQuestionChildren(ctx, db, rt, id); err != nil {
					return nil, err
				}

				for _, item := range model.Objects(vals) {
					item["question_id"] = id
					if _, err = repository.Write(ctx, db, rt, "", item); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if table == "practices" {
		if ids, ok := v["question_ids"]; ok {
			if _, err = repository.ClearPracticeQuestions(ctx, db, id); err != nil {
				return nil, err
			}

			seen := map[string]bool{}

			for i, qid := range model.Strings(ids) {
				if seen[qid] {
					continue
				}

				seen[qid] = true
				q, err := repository.Get(ctx, db, "questions", qid)
				if err != nil {
					return nil, err
				}

				if model.String(q, "review_status") != "approved" {
					return nil, fail(http.StatusUnprocessableEntity, "练习卷只能使用已审核题目")
				}

				if _, err = repository.Write(
					ctx,
					db,
					"practice_questions",
					"",
					model.Record{
						"practice_id": id,
						"question_id": qid,
						"sort_order":  i,
					},
				); err != nil {
					return nil, err
				}
			}
		}
	}

	return s.enrich(ctx, db, table, out)
}
