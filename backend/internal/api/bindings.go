package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"mixm-edu/backend/internal/store"
)

func (s *Server) bindingRoutes() {
	s.route("POST /api/bindings", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Bindings"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		if e = require(v, "content_id", "category_id"); e != nil {
			return nil, e
		}
		return store.Write(r.Context(), s.DB, "content_categories", "", v)
	})
	for _, p := range []string{"/api/bindings", "/api/bindings/content/{content_id}", "/api/bindings/category/{category_id}"} {
		s.route("GET "+p, func(w http.ResponseWriter, r *http.Request) (any, error) {
			query := "SELECT * FROM content_categories"
			args := []any{}
			for _, k := range []string{"content_id", "category_id"} {
				if id := r.PathValue(k); id != "" {
					query += " WHERE " + k + "=$1"
					args = append(args, id)
				}
			}
			return store.Query(r.Context(), s.DB, query+" ORDER BY sort_order,id", args...)
		})
	}
	s.route("GET /api/bindings/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return store.Get(r.Context(), s.DB, "content_categories", r.PathValue("id"))
	})
	s.route("DELETE /api/bindings/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Bindings"); e != nil {
			return nil, e
		}
		return store.Row{}, store.Delete(r.Context(), s.DB, "content_categories", r.PathValue("id"))
	})
	s.route("DELETE /api/bindings/by-content/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Bindings"); e != nil {
			return nil, e
		}
		_, e := s.DB.Exec(r.Context(), "DELETE FROM content_categories WHERE content_id=$1", r.PathValue("id"))
		return store.Row{}, e
	})
	s.route("GET /api/bindings/detail/content/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := store.Get(r.Context(), s.DB, "contents", r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		return s.enrich(r, s.DB, "contents", v)
	})
	s.route("GET /api/bindings/detail/category/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := store.Get(r.Context(), s.DB, "categories", r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		v["contents"], e = store.Query(r.Context(), s.DB, "SELECT c.* FROM contents c JOIN content_categories cc ON cc.content_id=c.id WHERE cc.category_id=$1 ORDER BY cc.sort_order", v["id"])
		return v, e
	})
	s.route("GET /api/bindings/category/{id}/unbound", func(w http.ResponseWriter, r *http.Request) (any, error) {
		rows, total, e := s.list(r, "contents", "NOT EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id=$1)", []any{r.PathValue("id")})
		return store.Row{"total": total, "contents": rows}, e
	})
	for _, a := range []string{"batch-bind", "batch-unbind"} {
		method := "POST"
		if a == "batch-unbind" {
			method = "DELETE"
		}
		s.route(method+" /api/bindings/category/{id}/"+a, func(w http.ResponseWriter, r *http.Request) (any, error) {
			if e := s.permission(r, "Bindings"); e != nil {
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
			ids := store.Strings(v["content_ids"])
			for _, id := range ids {
				if a == "batch-bind" {
					_, e = tx.Exec(r.Context(), "INSERT INTO content_categories(id,content_id,category_id) VALUES($1,$2,$3) ON CONFLICT(content_id,category_id) DO NOTHING", uuid.NewString(), id, r.PathValue("id"))
				} else {
					_, e = tx.Exec(r.Context(), "DELETE FROM content_categories WHERE category_id=$1 AND content_id=$2", r.PathValue("id"), id)
				}
				if e != nil {
					return nil, e
				}
			}
			if e = tx.Commit(r.Context()); e != nil {
				return nil, e
			}
			out := store.Row{"count": len(ids), "message": "操作成功"}
			respond(w, 200, out)
			return nil, nil
		})
	}
	s.route("POST /api/bindings/counts", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		out := store.Row{}
		ids := store.Strings(v["category_ids"])
		for _, id := range ids {
			out[id] = 0
		}
		rs, e := store.Query(r.Context(), s.DB, "SELECT category_id,count(*) AS n FROM content_categories WHERE category_id=ANY($1) GROUP BY category_id", ids)
		for _, v := range rs {
			out[store.String(v, "category_id")] = v["n"]
		}
		return out, e
	})
	s.route("GET /api/contents/by-category/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		cat, e := store.Get(r.Context(), s.DB, "categories", r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		rows, _, e := s.list(r, "contents", "is_active=true AND status='published' AND EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id IN (WITH RECURSIVE c AS (SELECT id FROM categories WHERE id=$1 UNION SELECT x.id FROM categories x JOIN c ON x.parent_id=c.id) SELECT id FROM c))", []any{cat["id"]})
		return store.Row{"category": cat, "contents": rows}, e
	})
	s.route("GET /api/contents/by-categories", func(w http.ResponseWriter, r *http.Request) (any, error) {
		ids := strings.Split(r.URL.Query().Get("category_ids"), ",")
		rows, total, e := s.list(r, "contents", "is_active=true AND status='published' AND EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id=ANY($1))", []any{ids})
		return store.Row{"total": total, "contents": rows}, e
	})
	s.route("GET /api/contents/category-tree/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		maxDepth := 3
		if text := r.URL.Query().Get("max_depth"); text != "" {
			n, err := strconv.Atoi(text)
			if err != nil || n < 1 || n > 20 {
				return nil, fail(422, "max_depth 必须为 1-20")
			}
			maxDepth = n
		}
		rs, e := store.Query(r.Context(), s.DB, "SELECT * FROM categories WHERE is_active=true ORDER BY sort_order,id")
		if e != nil {
			return nil, e
		}
		for _, row := range rs {
			if row["id"] == r.PathValue("id") {
				row["children"] = contentTree(rs, r.PathValue("id"), 2, maxDepth)
				row["type"] = "category"
				row["depth"] = 1
				return row, nil
			}
		}
		return nil, fail(404, "分类不存在")
	})
	s.route("GET /api/contents/categories-with-contents", func(w http.ResponseWriter, r *http.Request) (any, error) {
		cats, e := store.Query(r.Context(), s.DB, "SELECT * FROM categories WHERE parent_id IS NULL AND is_active=true ORDER BY sort_order,id")
		if e != nil {
			return nil, e
		}
		out := []store.Row{}
		for _, cat := range cats {
			rows, e := store.Query(r.Context(), s.DB, "SELECT c.* FROM contents c JOIN content_categories cc ON cc.content_id=c.id WHERE cc.category_id=$1 AND c.is_active=true AND c.status='published' ORDER BY cc.sort_order LIMIT 4", cat["id"])
			if e != nil {
				return nil, e
			}
			out = append(out, store.Row{"category": cat, "contents": rows})
		}
		return out, nil
	})
}

func contentTree(rows []store.Row, parent string, depth, max int) []store.Row {
	out := []store.Row{}
	if depth > max {
		return out
	}
	for _, row := range rows {
		if store.String(row, "parent_id") == parent {
			row["depth"], row["type"] = depth, "category"
			row["children"] = contentTree(rows, store.String(row, "id"), depth+1, max)
			out = append(out, row)
		}
	}
	return out
}
