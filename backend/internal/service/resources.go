package service

import (
	"context"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"

	"net/http"
	"net/url"
)

type ResourceSpec struct {
	Path, Table, Key, Permission string
	Array                        bool
}

var ResourcesList = []ResourceSpec{
	{Path: "categories", Table: "categories", Key: "categories", Permission: "Categories"}, {Path: "contents", Table: "contents", Key: "contents", Permission: "Contents"}, {Path: "content-types", Table: "content_types", Key: "types", Permission: "Contents"}, {Path: "templates", Table: "templates", Key: "templates", Permission: "Templates"}, {Path: "roles", Table: "roles", Key: "roles", Permission: "Roles"}, {Path: "questions", Table: "questions", Key: "questions", Permission: "Questions"}, {Path: "question-categories", Table: "question_categories", Key: "categories", Permission: "Questions"}, {Path: "practices", Table: "practices", Key: "practices", Permission: "Practices"}, {Path: "assignments", Table: "assignments", Key: "assignments", Permission: "Assignments"}}

type Resources struct{ *Core }

func validateResource(table, id string, v model.Record) error {
	for _, key := range []string{"parent_id", "name", "title", "type", "content_id"} {
		if value, ok := v[key]; ok && value != nil {
			if _, ok := value.(string); !ok {
				return fail(http.StatusUnprocessableEntity, key+" 必须为字符串")
			}
		}
	}
	for _, key := range []string{"visible_roles", "knowledge_tag_ids", "question_ids", "route_keys", "options"} {
		if value, ok := v[key]; ok && value != nil {
			switch value.(type) {
			case []string, []any, []model.Record:
			default:
				return fail(http.StatusUnprocessableEntity, key+" 必须为数组")
			}
		}
	}
	if id == "" {
		switch table {
		case "questions":
			if err := require(v, "title", "type"); err != nil {
				return err
			}
			if _, ok := v["difficulty"]; !ok {
				return fail(http.StatusUnprocessableEntity, "需要 difficulty")
			}
		case "contents", "assignments", "practices":
			if err := require(v, "title"); err != nil {
				return err
			}
		default:
			if err := require(v, "name"); err != nil {
				return err
			}
		}
	}
	if d, ok := v["difficulty"]; ok {
		n := model.Number(model.Record{"n": d}, "n")
		if n < 1 || n > 5 {
			return fail(http.StatusUnprocessableEntity, "difficulty 必须为 1-5")
		}
	}
	if table == "questions" {
		if x, ok := v["type"]; ok {
			valid := map[string]bool{"single_choice": true, "multi_choice": true, "true_false": true, "blank": true, "essay": true}
			kind, ok := x.(string)
			if !ok || !valid[kind] {
				return fail(http.StatusUnprocessableEntity, "题型无效")
			}
		}
	}
	return nil
}
func tree(rs []model.Record, parent string) []model.Record {
	out := []model.Record{}
	for _, row := range rs {
		if model.String(row, "parent_id") == parent {
			row["children"] = tree(rs, model.String(row, "id"))
			out = append(out, row)
		}
	}
	return out
}
func (s *Resources) categoryTree(ctx context.Context, table string) (any, error) {
	order := "id"
	if table == "categories" {
		order = "sort_order,id"
	}
	rs, err := repository.CategoryTreeRows(ctx, s.DB, order, table)
	if err != nil {
		return nil, err
	}
	return tree(rs, ""), nil
}
func (s *Resources) List(ctx context.Context, actor model.Actor, query url.Values, spec ResourceSpec) (any, error) {
	rs, total, err := s.list(ctx, actor, query, spec.Table, "", nil)
	if err != nil {
		return nil, err
	}
	for i, row := range rs {
		rs[i], err = s.enrich(ctx, s.DB, spec.Table, row)
		if err != nil {
			return nil, err
		}
	}

	if spec.Table == "categories" && query.Get("include") == "content-preview" {
		for _, cat := range rs {
			cat["contents"], err = repository.CategoryContentPreview(ctx, s.DB, model.String(cat, "id"))
			if err != nil {
				return nil, err
			}
		}
	}
	return model.Record{"total": total, spec.Key: rs}, nil
}
func (s *Resources) Detail(ctx context.Context, actor model.Actor, id string, query url.Values, spec ResourceSpec) (any, error) {
	v, err := repository.ResourceDetail(ctx, s.DB, actor, spec.Table, id)
	if err != nil {
		return nil, err
	}
	result, err := s.enrich(ctx, s.DB, spec.Table, v)
	if err != nil {
		return nil, err
	}
	if spec.Table == "categories" && query.Get("include") == "contents" {
		result["contents"], err = repository.BoundCategoryContents(ctx, s.DB, id)
	}
	return result, err
}
func (s *Resources) Save(ctx context.Context, actor model.Actor, input dto.ResourceInput, id string, spec ResourceSpec) (any, error) {
	if err := s.permission(ctx, actor, spec.Permission); err != nil {
		return nil, err
	}
	v := input.Fields()
	var err error
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	out, err := s.saveResource(ctx, actor, tx, spec.Table, id, v)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return out, nil
}
func (s *Resources) Delete(ctx context.Context, actor model.Actor, id string, spec ResourceSpec) (any, error) {
	if err := s.permission(ctx, actor, spec.Permission); err != nil {
		return nil, err
	}
	return model.Record{}, repository.Delete(ctx, s.DB, spec.Table, id)
}
func (s *Resources) CategoryTree(ctx context.Context) (any, error) {
	return s.categoryTree(ctx, "categories")
}
func (s *Resources) QuestionCategoryTree(ctx context.Context) (any, error) {
	return s.categoryTree(ctx, "question_categories")
}
func (s *Resources) BatchQuestions(ctx context.Context, actor model.Actor, inputs []dto.ResourceInput) (any, error) {
	if err := s.permission(ctx, actor, "Questions"); err != nil {
		return nil, err
	}
	items := inputs
	if len(items) > 500 {
		return nil, fail(http.StatusUnprocessableEntity, "一次最多 500 题")
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	out := []model.Record{}
	for _, v := range items {
		row, err := s.saveResource(ctx, actor, tx, "questions", "", v.Fields())
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, tx.Commit(ctx)
}
func (s *Resources) ReviewQuestion(ctx context.Context, actor model.Actor, input dto.ResourceInput, id string) (any, error) {
	if err := s.permission(ctx, actor, "Questions"); err != nil {
		return nil, err
	}
	v := input.Fields()
	var err error
	out, err := repository.ReviewQuestion(ctx, s.DB, model.String(v, "review_status"), actor.ID, id)
	if err != nil {
		return nil, err
	}
	return s.enrich(ctx, s.DB, "questions", out)
}
