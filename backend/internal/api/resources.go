package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mixm-edu/backend/internal/store"
)

type resourceSpec struct {
	Path, Table, Key, Permission string
	Array                        bool
}

var resources = []resourceSpec{{"categories", "categories", "categories", "Categories", true}, {"contents", "contents", "contents", "Contents", false}, {"content-types", "content_types", "types", "Contents", false}, {"templates", "templates", "templates", "Templates", false}, {"roles", "roles", "roles", "Roles", false}, {"questions", "questions", "questions", "Questions", false}, {"question-categories", "question_categories", "categories", "Questions", false}, {"practices", "practices", "practices", "Practices", false}, {"assignments", "assignments", "assignments", "Assignments", false}}

func page(r *http.Request) (int, int, error) {
	skip, limit := 0, 100
	var e error
	if x := r.URL.Query().Get("skip"); x != "" {
		skip, e = strconv.Atoi(x)
		if e != nil || skip < 0 {
			return 0, 0, fail(422, "skip 无效")
		}
	}
	if x := r.URL.Query().Get("limit"); x != "" {
		limit, e = strconv.Atoi(x)
		if e != nil || limit < 1 || limit > 1000 {
			return 0, 0, fail(422, "limit 必须为 1-1000")
		}
	}
	return skip, limit, nil
}
func (s *Server) visibility(r *http.Request, table string) string {
	if store.Bool(user(r), "is_superuser") {
		return "true"
	}
	singular := map[string]string{"assignments": "assignment", "questions": "question", "practices": "practice"}[table]
	if singular == "" {
		return "true"
	}
	id := singular + "_id"
	condition := fmt.Sprintf("(NOT EXISTS(SELECT 1 FROM %s_roles vr WHERE vr.%s=%s.id) OR EXISTS(SELECT 1 FROM %s_roles vr JOIN user_roles ur ON ur.role_id=vr.role_id WHERE vr.%s=%s.id AND ur.user_id='%s'))", singular, id, table, singular, id, table, strings.ReplaceAll(store.String(user(r), "id"), "'", "''"))
	if table == "questions" && s.permission(r, "Questions") != nil {
		condition += " AND review_status='approved'"
	}
	return condition
}
func (s *Server) list(r *http.Request, table, extra string, args []any) ([]store.Row, int, error) {
	skip, limit, e := page(r)
	if e != nil {
		return nil, 0, e
	}
	where := []string{s.visibility(r, table)}
	if extra != "" {
		where = append(where, extra)
	}
	q := r.URL.Query()
	for _, k := range []string{"parent_id", "content_type", "status", "template_id", "type", "review_status", "difficulty", "is_active", "is_ai_generated"} {
		if q.Has(k) {
			if _, ok := store.Columns[table][k]; ok {
				var filter any = q.Get(k)
				if store.Columns[table][k] == "integer" {
					n, err := strconv.Atoi(q.Get(k))
					if err != nil {
						return nil, 0, fail(422, k+" 无效")
					}
					filter = n
				}
				if store.Columns[table][k] == "boolean" {
					b, err := strconv.ParseBool(q.Get(k))
					if err != nil {
						return nil, 0, fail(422, k+" 无效")
					}
					filter = b
				}
				args = append(args, filter)
				if k == "parent_id" && q.Get(k) == "" {
					where = append(where, "parent_id IS NULL")
					args = args[:len(args)-1]
				} else {
					where = append(where, fmt.Sprintf("%s=$%d", k, len(args)))
				}
			}
		}
	}
	kw := q.Get("keyword")
	if kw == "" {
		kw = q.Get("search")
	}
	if kw != "" {
		field := "name"
		if _, ok := store.Columns[table]["title"]; ok {
			field = "title"
		}
		if table == "users" {
			field = "username"
		}
		args = append(args, "%"+kw+"%")
		where = append(where, fmt.Sprintf("%s ILIKE $%d", field, len(args)))
	}
	if q.Get("knowledge_tag_ids") != "" && (table == "questions" || table == "practices") {
		prefix := strings.TrimSuffix(table, "s")
		args = append(args, strings.Split(q.Get("knowledge_tag_ids"), ","))
		where = append(where, fmt.Sprintf("EXISTS(SELECT 1 FROM %s_tags t WHERE t.%s_id=%s.id AND t.category_id=ANY($%d))", prefix, prefix, table, len(args)))
	}
	clause := strings.Join(where, " AND ")
	count, e := store.One(r.Context(), s.DB, "SELECT count(*) AS total FROM "+table+" WHERE "+clause, args...)
	if e != nil {
		return nil, 0, e
	}
	order := "id"
	if _, ok := store.Columns[table]["created_at"]; ok {
		order = "created_at DESC,id"
	}
	if table == "categories" {
		order = "sort_order,id"
	}
	args = append(args, limit, skip)
	rows, e := store.Query(r.Context(), s.DB, fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d", table, clause, order, len(args)-1, len(args)), args...)
	return rows, int(store.Number(count, "total")), e
}
func (s *Server) enrich(r *http.Request, db store.DB, table string, v store.Row) (store.Row, error) {
	id := v["id"]
	var e error
	prefix := map[string]string{"questions": "question", "practices": "practice", "assignments": "assignment"}[table]
	if prefix != "" {
		roles, err := store.Query(r.Context(), db, "SELECT role_id FROM "+prefix+"_roles WHERE "+prefix+"_id=$1", id)
		if err != nil {
			return nil, err
		}
		ids := []string{}
		for _, role := range roles {
			ids = append(ids, store.String(role, "role_id"))
		}
		v["visible_roles"] = ids
	}
	if table == "questions" || table == "practices" {
		tags, err := store.Query(r.Context(), db, "SELECT c.id,c.name FROM categories c JOIN "+prefix+"_tags t ON t.category_id=c.id WHERE t."+prefix+"_id=$1", id)
		if err != nil {
			return nil, err
		}
		ids := []string{}
		for _, t := range tags {
			ids = append(ids, store.String(t, "id"))
		}
		v["knowledge_tags"], v["knowledge_tag_ids"] = tags, ids
	}
	switch table {
	case "questions":
		v["options"], e = store.Query(r.Context(), db, "SELECT * FROM question_options WHERE question_id=$1 ORDER BY option_key", id)
		if e != nil {
			return nil, e
		}
		v["attachments"], e = store.Query(r.Context(), db, "SELECT * FROM question_attachments WHERE question_id=$1", id)
	case "practices":
		rows, err := store.Query(r.Context(), db, "SELECT q.* FROM questions q JOIN practice_questions pq ON pq.question_id=q.id WHERE pq.practice_id=$1 ORDER BY pq.sort_order", id)
		if err != nil {
			return nil, err
		}
		for i, q := range rows {
			rows[i], err = s.enrich(r, db, "questions", q)
			if err != nil {
				return nil, err
			}
		}
		v["questions"], v["question_count"] = rows, len(rows)
	case "contents":
		v["categories"], e = store.Query(r.Context(), db, "SELECT c.* FROM categories c JOIN content_categories cc ON cc.category_id=c.id WHERE cc.content_id=$1 ORDER BY cc.sort_order", id)
	case "assignments":
		v["attachments"] = store.Objects(v["attachments"])
		cnt, err := store.One(r.Context(), db, "SELECT count(*) AS n FROM submissions WHERE assignment_id=$1", id)
		if err != nil {
			return nil, err
		}
		v["submission_count"] = cnt["n"]
	}
	return v, e
}
func validateResource(table, id string, v store.Row) error {
	for _, key := range []string{"parent_id", "name", "title", "type", "content_id"} {
		if value, ok := v[key]; ok && value != nil {
			if _, ok := value.(string); !ok {
				return fail(422, key+" 必须为字符串")
			}
		}
	}
	for _, key := range []string{"visible_roles", "knowledge_tag_ids", "question_ids", "route_keys", "options"} {
		if value, ok := v[key]; ok && value != nil {
			switch value.(type) {
			case []string, []any, []store.Row:
			default:
				return fail(422, key+" 必须为数组")
			}
		}
	}
	if id == "" {
		switch table {
		case "questions":
			if e := require(v, "title", "type"); e != nil {
				return e
			}
			if _, ok := v["difficulty"]; !ok {
				return fail(422, "需要 difficulty")
			}
		case "contents", "assignments", "practices":
			if e := require(v, "title"); e != nil {
				return e
			}
		default:
			if e := require(v, "name"); e != nil {
				return e
			}
		}
	}
	if d, ok := v["difficulty"]; ok {
		n := store.Number(store.Row{"n": d}, "n")
		if n < 1 || n > 5 {
			return fail(422, "difficulty 必须为 1-5")
		}
	}
	if table == "questions" {
		if x, ok := v["type"]; ok {
			valid := map[string]bool{"single_choice": true, "multi_choice": true, "true_false": true, "blank": true, "essay": true}
			kind, ok := x.(string)
			if !ok || !valid[kind] {
				return fail(422, "题型无效")
			}
		}
	}
	return nil
}
func (s *Server) saveResource(r *http.Request, db store.DB, table, id string, v store.Row) (store.Row, error) {
	if e := validateResource(table, id, v); e != nil {
		return nil, e
	}
	if id == "" {
		if _, ok := store.Columns[table]["created_by"]; ok {
			v["created_by"] = user(r)["id"]
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
				v["config"] = store.JSON(v["config"])
			}
		}
	}
	if table == "assignments" {
		if _, ok := v["attachments"]; ok {
			v["attachments"] = store.JSON(v["attachments"])
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
					return nil, fail(422, "分类不能作为自己的父级")
				}
				parent, e := store.Get(r.Context(), db, table, fmt.Sprint(p))
				if e != nil {
					return nil, e
				}
				if id != "" {
					cycle, e := store.One(r.Context(), db, "WITH RECURSIVE c AS (SELECT id,parent_id FROM "+table+" WHERE id=$1 UNION SELECT x.id,x.parent_id FROM "+table+" x JOIN c ON x.id=c.parent_id) SELECT EXISTS(SELECT 1 FROM c WHERE id=$2) AS cycle", p, id)
					if e != nil {
						return nil, e
					}
					if store.Bool(cycle, "cycle") {
						return nil, fail(422, "分类形成循环")
					}
				}
				v["level"] = int(store.Number(parent, "level")) + 1
			}
		}
	}
	out, e := store.Write(r.Context(), db, table, id, v)
	if e != nil {
		return nil, e
	}
	id = store.String(out, "id")
	prefix := map[string]string{"questions": "question", "practices": "practice", "assignments": "assignment"}[table]
	for _, rel := range []struct{ field, suffix, col string }{{"visible_roles", "roles", "role_id"}, {"knowledge_tag_ids", "tags", "category_id"}} {
		if vals, ok := v[rel.field]; ok && prefix != "" && !(table == "assignments" && rel.suffix == "tags") {
			rt := prefix + "_" + rel.suffix
			if _, e = db.Exec(r.Context(), "DELETE FROM "+rt+" WHERE "+prefix+"_id=$1", id); e != nil {
				return nil, e
			}
			seen := map[string]bool{}
			for _, rid := range store.Strings(vals) {
				if seen[rid] {
					continue
				}
				seen[rid] = true
				if _, e = db.Exec(r.Context(), "INSERT INTO "+rt+" ("+prefix+"_id,"+rel.col+") VALUES($1,$2)", id, rid); e != nil {
					return nil, e
				}
			}
		}
	}
	if table == "questions" {
		for _, field := range []string{"options", "attachments"} {
			if vals, ok := v[field]; ok {
				rt := "question_" + field
				if _, e = db.Exec(r.Context(), "DELETE FROM "+rt+" WHERE question_id=$1", id); e != nil {
					return nil, e
				}
				for _, item := range store.Objects(vals) {
					item["question_id"] = id
					if _, e = store.Write(r.Context(), db, rt, "", item); e != nil {
						return nil, e
					}
				}
			}
		}
	}
	if table == "practices" {
		if ids, ok := v["question_ids"]; ok {
			if _, e = db.Exec(r.Context(), "DELETE FROM practice_questions WHERE practice_id=$1", id); e != nil {
				return nil, e
			}
			seen := map[string]bool{}
			for i, qid := range store.Strings(ids) {
				if seen[qid] {
					continue
				}
				seen[qid] = true
				q, e := store.Get(r.Context(), db, "questions", qid)
				if e != nil {
					return nil, e
				}
				if store.String(q, "review_status") != "approved" {
					return nil, fail(422, "练习卷只能使用已审核题目")
				}
				if _, e = store.Write(r.Context(), db, "practice_questions", "", store.Row{"practice_id": id, "question_id": qid, "sort_order": i}); e != nil {
					return nil, e
				}
			}
		}
	}
	return s.enrich(r, db, table, out)
}
func (s *Server) resourceRoutes() {
	for _, res := range resources {
		base := "/api/" + res.Path
		s.route("GET "+base, func(w http.ResponseWriter, r *http.Request) (any, error) {
			rs, total, e := s.list(r, res.Table, "", nil)
			if e != nil {
				return nil, e
			}
			for i, row := range rs {
				rs[i], e = s.enrich(r, s.DB, res.Table, row)
				if e != nil {
					return nil, e
				}
			}
			if res.Array {
				return rs, nil
			}
			return store.Row{"total": total, res.Key: rs}, nil
		})
		s.route("GET "+base+"/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
			v, e := store.One(r.Context(), s.DB, "SELECT * FROM "+res.Table+" WHERE id=$1 AND "+s.visibility(r, res.Table), r.PathValue("id"))
			if e != nil {
				return nil, e
			}
			return s.enrich(r, s.DB, res.Table, v)
		})
		for _, method := range []string{"POST", "PUT"} {
			path := base
			if method == "PUT" {
				path += "/{id}"
			}
			s.route(method+" "+path, func(w http.ResponseWriter, r *http.Request) (any, error) {
				if e := s.permission(r, res.Permission); e != nil {
					return nil, e
				}
				v, e := body(r)
				if e != nil {
					return nil, e
				}
				tx, e := s.DB.Begin(r.Context())
				if e != nil {
					return nil, e
				}
				defer tx.Rollback(r.Context())
				out, e := s.saveResource(r, tx, res.Table, r.PathValue("id"), v)
				if e != nil {
					return nil, e
				}
				if e = tx.Commit(r.Context()); e != nil {
					return nil, e
				}
				if r.Method == "POST" {
					respond(w, 201, out)
					return nil, nil
				}
				return out, nil
			})
		}
		s.route("DELETE "+base+"/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
			if e := s.permission(r, res.Permission); e != nil {
				return nil, e
			}
			return store.Row{}, store.Delete(r.Context(), s.DB, res.Table, r.PathValue("id"))
		})
	}
	s.route("GET /api/categories/tree", s.categoryTree("categories"))
	s.route("GET /api/question-categories/tree", s.categoryTree("question_categories"))
	s.route("POST /api/questions/batch", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Questions"); e != nil {
			return nil, e
		}
		var items []store.Row
		if e := decode(r, &items); e != nil {
			return nil, e
		}
		if len(items) > 500 {
			return nil, fail(422, "一次最多 500 题")
		}
		tx, e := s.DB.Begin(r.Context())
		if e != nil {
			return nil, e
		}
		defer tx.Rollback(r.Context())
		out := []store.Row{}
		for _, v := range items {
			row, e := s.saveResource(r, tx, "questions", "", v)
			if e != nil {
				return nil, e
			}
			out = append(out, row)
		}
		return out, tx.Commit(r.Context())
	})
	s.route("POST /api/questions/{id}/review", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Questions"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		out, e := store.One(r.Context(), s.DB, "UPDATE questions SET review_status=$1,reviewed_by=$2,reviewed_at=now() WHERE id=$3 RETURNING *", v["review_status"], user(r)["id"], r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		return s.enrich(r, s.DB, "questions", out)
	})
	s.route("GET /api/questions/ai-generated", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Questions"); e != nil {
			return nil, e
		}
		rs, total, e := s.list(r, "questions", "is_ai_generated=true", nil)
		if e != nil {
			return nil, e
		}
		for i, v := range rs {
			rs[i], e = s.enrich(r, s.DB, "questions", v)
			if e != nil {
				return nil, e
			}
		}
		return store.Row{"total": total, "questions": rs}, nil
	})
}
func (s *Server) categoryTree(table string) handler {
	return func(w http.ResponseWriter, r *http.Request) (any, error) {
		order := "id"
		if table == "categories" {
			order = "sort_order,id"
		}
		rs, e := store.Query(r.Context(), s.DB, "SELECT * FROM "+table+" ORDER BY "+order)
		if e != nil {
			return nil, e
		}
		return tree(rs, ""), nil
	}
}
func tree(rs []store.Row, parent string) []store.Row {
	out := []store.Row{}
	for _, row := range rs {
		if store.String(row, "parent_id") == parent {
			row["children"] = tree(rs, store.String(row, "id"))
			out = append(out, row)
		}
	}
	return out
}
