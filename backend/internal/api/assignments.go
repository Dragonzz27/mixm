package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"mixm-edu/backend/internal/store"
)

func submissionResponse(v store.Row) store.Row {
	if p := store.String(v, "file_path"); p != "" {
		v["file_url"] = "/api/assignments/file/" + filepath.Base(p)
	}
	return v
}
func (s *Server) assignmentRoutes() {
	s.route("GET /api/assignments/active", func(w http.ResponseWriter, r *http.Request) (any, error) {
		rs, _, e := s.list(r, "assignments", "status='active' AND deadline>now()", nil)
		if e != nil {
			return nil, e
		}
		for i, v := range rs {
			rs[i], e = s.enrich(r, s.DB, "assignments", v)
			if e != nil {
				return nil, e
			}
		}
		return rs, nil
	})
	s.route("GET /api/assignments/my/submissions", func(w http.ResponseWriter, r *http.Request) (any, error) {
		query := "SELECT * FROM submissions WHERE user_id=$1"
		args := []any{user(r)["id"]}
		if id := r.URL.Query().Get("assignment_id"); id != "" {
			query += " AND assignment_id=$2"
			args = append(args, id)
		}
		rs, e := store.Query(r.Context(), s.DB, query+" ORDER BY submitted_at DESC", args...)
		for i, v := range rs {
			rs[i] = submissionResponse(v)
		}
		return rs, e
	})
	s.route("POST /api/assignments/submissions", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		a, e := store.One(r.Context(), s.DB, "SELECT * FROM assignments WHERE id=$1 AND "+s.visibility(r, "assignments"), v["assignment_id"])
		if e != nil {
			return nil, e
		}
		deadline, _ := a["deadline"].(time.Time)
		if store.String(a, "status") != "active" || deadline.Before(time.Now()) {
			return nil, fail(400, "作业已截止")
		}
		v["user_id"], v["status"] = user(r)["id"], "submitted"
		delete(v, "grade")
		delete(v, "feedback")
		delete(v, "reviewed_at")
		if p := store.String(v, "file_path"); p != "" {
			if filepath.Clean(p) != filepath.Join(s.Config.UploadDir, filepath.Base(p)) {
				return nil, fail(422, "文件路径无效")
			}
			if _, e = os.Stat(p); e != nil {
				return nil, fail(422, "提交文件不存在")
			}
		}
		if v["content"] == nil && v["file_path"] == nil {
			return nil, fail(422, "需要提交内容或文件")
		}
		row, e := store.Write(r.Context(), s.DB, "submissions", "", v)
		if e != nil {
			return nil, e
		}
		respond(w, 201, submissionResponse(row))
		return nil, nil
	})
	s.route("GET /api/assignments/{id}/submissions", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Assignments"); e != nil {
			return nil, e
		}
		rs, total, e := s.list(r, "submissions", "assignment_id=$1", []any{r.PathValue("id")})
		for i, v := range rs {
			rs[i] = submissionResponse(v)
		}
		return store.Row{"total": total, "submissions": rs}, e
	})
	s.route("PUT /api/assignments/submissions/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Assignments"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		allowed := store.Row{"reviewed_at": time.Now(), "status": "reviewed"}
		for _, k := range []string{"grade", "feedback", "status"} {
			if x, ok := v[k]; ok {
				allowed[k] = x
			}
		}
		out, e := store.Write(r.Context(), s.DB, "submissions", r.PathValue("id"), allowed)
		if e != nil {
			return nil, e
		}
		return submissionResponse(out), nil
	})
	s.route("DELETE /api/assignments/submissions/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Assignments"); e != nil {
			return nil, e
		}
		return store.Row{}, store.Delete(r.Context(), s.DB, "submissions", r.PathValue("id"))
	})
	s.route("GET /api/assignments/{id}/submissions/export", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Assignments"); e != nil {
			return nil, e
		}
		rows, e := store.Query(r.Context(), s.DB, "SELECT s.*,u.username FROM submissions s JOIN users u ON u.id=s.user_id WHERE s.assignment_id=$1 ORDER BY s.submitted_at", r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		if e = os.MkdirAll(s.Config.ExportDir, 0755); e != nil {
			return nil, e
		}
		name := "grades-" + uuid.NewString() + ".csv"
		f, e := os.Create(filepath.Join(s.Config.ExportDir, name))
		if e != nil {
			return nil, e
		}
		defer f.Close()
		writer := csv.NewWriter(f)
		keys := []string{"id", "user_id", "username", "file_name", "submitted_at", "status", "grade", "feedback"}
		if e = writer.Write(keys); e != nil {
			return nil, e
		}
		for _, v := range rows {
			line := []string{}
			for _, k := range keys {
				text := ""
				if v[k] != nil {
					text = fmt.Sprint(v[k])
				}
				if len(text) > 0 && (text[0] == '=' || text[0] == '+' || text[0] == '-' || text[0] == '@') {
					text = "'" + text
				}
				line = append(line, text)
			}
			if e = writer.Write(line); e != nil {
				return nil, e
			}
		}
		writer.Flush()
		if e = writer.Error(); e != nil {
			return nil, e
		}
		return store.Row{"file_url": "/static/export/" + name, "filename": name}, nil
	})
}
