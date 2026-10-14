package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xuri/excelize/v2"
	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/store"
)

//go:embed navigation.json
var navigation []byte

type navData struct {
	Routes   []store.Row         `json:"routes"`
	Menu     []store.Row         `json:"menu"`
	Defaults []string            `json:"default_client_menu_keys"`
	Names    map[string][]string `json:"menu_key_route_names"`
}

func navigationFor(u store.Row, keys []string) ([]store.Row, []store.Row) {
	var n navData
	_ = json.Unmarshal(navigation, &n)
	if store.Bool(u, "is_superuser") {
		return n.Routes, n.Menu
	}
	allowed, names := map[string]bool{}, map[string]bool{}
	for _, k := range keys {
		allowed[k] = true
		for _, name := range n.Names[k] {
			names[name] = true
		}
	}
	routes, menu := []store.Row{}, []store.Row{}
	for _, m := range n.Menu {
		if allowed[store.String(m, "key")] {
			menu = append(menu, m)
		}
	}
	for _, r := range n.Routes {
		path := store.String(r, "path")
		if path == "/login" || path == "/register" || path == "/:pathMatch(.*)*" {
			routes = append(routes, r)
			continue
		}
		children := []store.Row{}
		for _, ch := range store.Objects(r["children"]) {
			if names[store.String(ch, "name")] {
				children = append(children, ch)
			}
		}
		if len(children) > 0 {
			r["children"] = children
			routes = append(routes, r)
		}
	}
	return routes, menu
}
func (s *Server) serializeUser(r *http.Request, u store.Row) (store.Row, error) {
	delete(u, "hashed_password")
	roles, e := store.Query(r.Context(), s.DB, "SELECT r.id,r.name,r.route_keys FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1", u["id"])
	if e != nil {
		return nil, e
	}
	u["role"], u["role_id"] = nil, nil
	if len(roles) > 0 {
		u["role"] = store.Row{"id": roles[0]["id"], "name": roles[0]["name"]}
		u["role_id"] = roles[0]["id"]
	}
	return u, nil
}
func (s *Server) tokens(u store.Row) (store.Row, error) {
	a, e := auth.Sign(s.Config.Secret, store.String(u, "username"), "access", s.Config.AccessTTL)
	if e != nil {
		return nil, e
	}
	b, e := auth.Sign(s.Config.Secret, store.String(u, "username"), "refresh", s.Config.RefreshTTL)
	return store.Row{"access_token": a, "refresh_token": b, "token_type": "bearer"}, e
}
func (s *Server) permission(r *http.Request, key string) error {
	if store.Bool(user(r), "is_superuser") {
		return nil
	}
	v, e := store.One(r.Context(), s.DB, "SELECT EXISTS(SELECT 1 FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1 AND r.route_keys ? $2) AS allowed", user(r)["id"], key)
	if e != nil {
		return e
	}
	if !store.Bool(v, "allowed") {
		return fail(403, "没有操作权限")
	}
	return nil
}
func (s *Server) saveUser(r *http.Request, id string, v store.Row) (store.Row, error) {
	if id == "" {
		if e := require(v, "username", "password"); e != nil {
			return nil, e
		}
		if store.String(v, "email") == "" && store.String(v, "phone") == "" {
			return nil, fail(422, "需要邮箱或手机号码")
		}
		hash, e := auth.Hash(store.String(v, "password"))
		if e != nil {
			return nil, fail(422, "密码长度无效")
		}
		v["hashed_password"] = hash
	}
	delete(v, "password")
	delete(v, "is_superuser")
	if v["email"] == "" {
		v["email"] = nil
	}
	if v["phone"] == "" {
		v["phone"] = nil
	}
	tx, e := s.DB.Begin(r.Context())
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(r.Context())
	u, e := store.Write(r.Context(), tx, "users", id, v)
	if e != nil {
		return nil, e
	}
	if role, ok := v["role_id"]; ok {
		if _, e = tx.Exec(r.Context(), "DELETE FROM user_roles WHERE user_id=$1", u["id"]); e != nil {
			return nil, e
		}
		if role != nil && role != "" {
			if _, e = store.Write(r.Context(), tx, "user_roles", "", store.Row{"user_id": u["id"], "role_id": role}); e != nil {
				return nil, e
			}
		}
	}
	if e = tx.Commit(r.Context()); e != nil {
		return nil, e
	}
	return s.serializeUser(r, u)
}
func (s *Server) authRoutes() {
	s.route("POST /api/auth/register", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Users"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		return s.saveUser(r, "", v)
	})
	s.route("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		u, e := store.One(r.Context(), s.DB, "SELECT * FROM users WHERE email=$1 AND is_active=true", v["email"])
		if e != nil || !auth.Verify(store.String(u, "hashed_password"), store.String(v, "password")) {
			return nil, fail(401, "邮箱或密码错误")
		}
		tokens, e := s.tokens(u)
		if e != nil {
			return nil, e
		}
		roles, e := store.Query(r.Context(), s.DB, "SELECT r.* FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1", u["id"])
		if e != nil {
			return nil, e
		}
		keys := []string{}
		if len(roles) > 0 {
			keys = store.Strings(roles[0]["route_keys"])
		}
		if len(keys) == 0 {
			keys = []string{"ClientHome", "ClientContents", "ClientMyInstances"}
		}
		routes, menu := navigationFor(u, keys)
		u, e = s.serializeUser(r, u)
		if e != nil {
			return nil, e
		}
		u["user_role"] = u["role"]
		u["role"] = "client"
		if store.Bool(u, "is_superuser") {
			u["role"] = "admin"
		}
		tokens["user"], tokens["routes"], tokens["menu"] = u, routes, menu
		return tokens, nil
	}, true)
	s.route("POST /api/auth/refresh", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		c, e := auth.Parse(s.Config.Secret, store.String(v, "refresh_token"), "refresh")
		if e != nil {
			return nil, fail(401, "刷新令牌无效")
		}
		u, e := store.One(r.Context(), s.DB, "SELECT * FROM users WHERE username=$1 AND is_active=true", c.Subject)
		if e != nil {
			return nil, fail(401, "用户不可用")
		}
		return s.tokens(u)
	}, true)
	s.route("GET /api/auth/users", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Users"); e != nil {
			return nil, e
		}
		rs, total, e := s.list(r, "users", "", nil)
		if e != nil {
			return nil, e
		}
		for i, u := range rs {
			rs[i], e = s.serializeUser(r, u)
			if e != nil {
				return nil, e
			}
		}
		return store.Row{"total": total, "users": rs}, nil
	})
	s.route("GET /api/auth/users/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if r.PathValue("id") != store.String(user(r), "id") {
			if e := s.permission(r, "Users"); e != nil {
				return nil, e
			}
		}
		u, e := store.Get(r.Context(), s.DB, "users", r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		return s.serializeUser(r, u)
	})
	s.route("PUT /api/auth/users/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Users"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		return s.saveUser(r, r.PathValue("id"), v)
	})
	s.route("POST /api/auth/users/{id}/change-password", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Users"); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		if e = require(v, "new_password"); e != nil {
			return nil, e
		}
		h, e := auth.Hash(store.String(v, "new_password"))
		if e != nil {
			return nil, fail(422, "密码长度无效")
		}
		_, e = store.Write(r.Context(), s.DB, "users", r.PathValue("id"), store.Row{"hashed_password": h})
		return store.Row{"message": "密码修改成功"}, e
	})
	s.route("DELETE /api/auth/users/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Users"); e != nil {
			return nil, e
		}
		if r.PathValue("id") == store.String(user(r), "id") {
			return nil, fail(400, "不能删除当前用户")
		}
		return store.Row{}, store.Delete(r.Context(), s.DB, "users", r.PathValue("id"))
	})
	s.route("POST /api/auth/users/import", s.importUsers)
	s.route("GET /api/roles/permissions", func(w http.ResponseWriter, r *http.Request) (any, error) {
		var n navData
		_ = json.Unmarshal(navigation, &n)
		out := []store.Row{}
		for _, m := range n.Menu {
			if store.String(m, "key") != "" && store.String(m, "type") != "divider" {
				out = append(out, store.Row{"key": m["key"], "label": m["label"]})
			}
		}
		return out, nil
	})
}
func (s *Server) importUsers(w http.ResponseWriter, r *http.Request) (any, error) {
	if e := s.permission(r, "Users"); e != nil {
		return nil, e
	}
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
	if e := r.ParseMultipartForm(20 << 20); e != nil {
		return nil, fail(422, "上传文件无效")
	}
	defer r.MultipartForm.RemoveAll()
	f, _, e := r.FormFile("file")
	if e != nil {
		return nil, fail(422, "需要 file 字段")
	}
	defer f.Close()
	book, e := excelize.OpenReader(f)
	if e != nil {
		return nil, fail(422, "需要 Excel 文件")
	}
	defer book.Close()
	names := book.GetSheetList()
	if len(names) == 0 {
		return nil, fail(422, "Excel 没有工作表")
	}
	rows, e := book.GetRows(names[0])
	if e != nil {
		return nil, e
	}
	if len(rows) == 0 {
		return nil, fail(422, "文件为空")
	}
	headers := map[string]int{}
	for i, h := range rows[0] {
		key := strings.ToLower(strings.TrimSpace(h))
		if key == "role" || key == "角色名" {
			key = "role_name"
		}
		headers[key] = i
	}
	if _, ok := headers["username"]; !ok {
		return nil, fail(422, "需要 username 列")
	}
	roleRows, e := store.Query(r.Context(), s.DB, "SELECT id,name FROM roles")
	if e != nil {
		return nil, e
	}
	roleIDs := map[string]string{}
	for _, role := range roleRows {
		roleIDs[store.String(role, "name")] = store.String(role, "id")
	}
	out := store.Row{"total": len(rows) - 1, "success": 0, "failed": 0}
	errs := []string{}
	success := 0
	for i, row := range rows[1:] {
		if strings.TrimSpace(strings.Join(row, "")) == "" {
			continue
		}
		v := store.Row{}
		for _, key := range []string{"username", "email", "phone", "password", "role_id", "role_name"} {
			if col, ok := headers[key]; ok && col < len(row) {
				v[key] = strings.TrimSpace(row[col])
			}
		}
		if store.String(v, "password") == "" {
			v["password"] = "123456"
		}
		if store.String(v, "role_id") == "" {
			id := roleIDs[store.String(v, "role_name")]
			if id == "" {
				errs = append(errs, fmt.Sprintf("第 %d 行: 角色不存在", i+2))
				continue
			}
			v["role_id"] = id
		}
		if _, e = s.saveUser(r, "", v); e != nil {
			errs = append(errs, fmt.Sprintf("第 %d 行: %s", i+2, e))
		} else {
			success++
		}
	}
	out["success"], out["failed"], out["errors"] = success, len(errs), errs
	out["total"] = success + len(errs)
	return out, nil
}
