package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/xuri/excelize/v2"

	"mixm/backend/internal/auth"
	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

//go:embed navigation.json
var navigation []byte

type navData struct {
	Routes   []model.Record      `json:"routes"`
	Menu     []model.Record      `json:"menu"`
	Defaults []string            `json:"default_client_menu_keys"`
	Names    map[string][]string `json:"menu_key_route_names"`
}

type Users struct{ *Core }

func navigationFor(u model.Record, keys []string) ([]model.Record, []model.Record) {
	var n navData
	_ = json.Unmarshal(navigation, &n)
	if model.Bool(u, "is_superuser") {
		return n.Routes, n.Menu
	}
	allowed, names := map[string]bool{}, map[string]bool{}
	for _, k := range keys {
		allowed[k] = true
		for _, name := range n.Names[k] {
			names[name] = true
		}
	}
	routes, menu := []model.Record{}, []model.Record{}
	for _, m := range n.Menu {
		if allowed[model.String(m, "key")] {
			menu = append(menu, m)
		}
	}
	for _, r := range n.Routes {
		path := model.String(r, "path")
		if path == "/login" || path == "/register" || path == "/:pathMatch(.*)*" {
			routes = append(routes, r)
			continue
		}
		children := []model.Record{}
		for _, ch := range model.Objects(r["children"]) {
			if names[model.String(ch, "name")] {
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
func (s *Users) importUsers(ctx context.Context, actor model.Actor, upload dto.Upload) (any, error) {
	if err := s.permission(ctx, actor, "Users"); err != nil {
		return nil, err
	}
	f := upload.Reader
	var err error

	book, err := excelize.OpenReader(f)
	if err != nil {
		return nil, fail(http.StatusUnprocessableEntity, "需要 Excel 文件")
	}
	defer book.Close()
	names := book.GetSheetList()
	if len(names) == 0 {
		return nil, fail(http.StatusUnprocessableEntity, "Excel 没有工作表")
	}
	rows, err := book.GetRows(names[0])
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fail(http.StatusUnprocessableEntity, "文件为空")
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
		return nil, fail(http.StatusUnprocessableEntity, "需要 username 列")
	}
	roleRows, err := repository.AllRoles(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	roleIDs := map[string]string{}
	for _, role := range roleRows {
		roleIDs[model.String(role, "name")] = model.String(role, "id")
	}
	out := model.Record{"total": len(rows) - 1, "success": 0, "failed": 0}
	errs := []string{}
	success := 0
	for i, row := range rows[1:] {
		if strings.TrimSpace(strings.Join(row, "")) == "" {
			continue
		}
		v := model.Record{}
		for _, key := range []string{"username", "email", "phone", "password", "role_id", "role_name"} {
			if col, ok := headers[key]; ok && col < len(row) {
				v[key] = strings.TrimSpace(row[col])
			}
		}
		if model.String(v, "password") == "" {
			v["password"] = "123456"
		}
		if model.String(v, "role_id") == "" {
			id := roleIDs[model.String(v, "role_name")]
			if id == "" {
				errs = append(errs, fmt.Sprintf("第 %d 行: 角色不存在", i+2))
				continue
			}
			v["role_id"] = id
		}
		if _, err = s.saveUser(ctx, "", v); err != nil {
			errs = append(errs, fmt.Sprintf("第 %d 行: %s", i+2, err))
		} else {
			success++
		}
	}
	out["success"], out["failed"], out["errors"] = success, len(errs), errs
	out["total"] = success + len(errs)
	return out, nil
}
func (s *Users) Create(ctx context.Context, actor model.Actor, input dto.UserInput) (any, error) {
	if err := s.permission(ctx, actor, "Users"); err != nil {
		return nil, err
	}
	v := input.Fields()
	return s.saveUser(ctx, "", v)
}
func (s *Users) Login(ctx context.Context, input dto.UserInput) (any, error) {
	v := input.Fields()
	var err error
	u, err := repository.LoginUser(ctx, s.DB, model.String(v, "email"))
	if err != nil || !auth.Verify(model.String(u, "hashed_password"), model.String(v, "password")) {
		return nil, fail(http.StatusUnauthorized, "邮箱或密码错误")
	}
	tokens, err := s.tokens(u)
	if err != nil {
		return nil, err
	}
	roles, err := repository.UserRoles(ctx, s.DB, model.String(u, "id"))
	if err != nil {
		return nil, err
	}
	keys := []string{}
	if len(roles) > 0 {
		keys = model.Strings(roles[0]["route_keys"])
	}
	if len(keys) == 0 {
		keys = []string{"ClientHome", "ClientContents", "ClientMyInstances"}
	}
	routes, menu := navigationFor(u, keys)
	u, err = s.serializeUser(ctx, u)
	if err != nil {
		return nil, err
	}
	u["user_role"] = u["role"]
	u["role"] = "client"
	if model.Bool(u, "is_superuser") {
		u["role"] = "admin"
	}
	tokens["user"], tokens["routes"], tokens["menu"] = u, routes, menu
	return tokens, nil
}
func (s *Users) Refresh(ctx context.Context, input dto.UserInput) (any, error) {
	v := input.Fields()
	var err error
	c, err := auth.Parse(s.Config.Secret, model.String(v, "refresh_token"), "refresh")
	if err != nil {
		return nil, fail(http.StatusUnauthorized, "刷新令牌无效")
	}
	u, err := repository.RefreshUser(ctx, s.DB, c.Subject)
	if err != nil {
		return nil, fail(http.StatusUnauthorized, "用户不可用")
	}
	return s.tokens(u)
}
func (s *Users) List(ctx context.Context, actor model.Actor, query url.Values) (any, error) {
	if err := s.permission(ctx, actor, "Users"); err != nil {
		return nil, err
	}
	rs, total, err := s.list(ctx, actor, query, "users", "", nil)
	if err != nil {
		return nil, err
	}
	for i, u := range rs {
		rs[i], err = s.serializeUser(ctx, u)
		if err != nil {
			return nil, err
		}
	}
	return model.Record{"total": total, "users": rs}, nil
}
func (s *Users) Detail(ctx context.Context, actor model.Actor, id string) (any, error) {
	if id != actor.ID {
		if err := s.permission(ctx, actor, "Users"); err != nil {
			return nil, err
		}
	}
	u, err := repository.Get(ctx, s.DB, "users", id)
	if err != nil {
		return nil, err
	}
	return s.serializeUser(ctx, u)
}
func (s *Users) Update(ctx context.Context, actor model.Actor, input dto.UserInput, id string) (any, error) {
	if err := s.permission(ctx, actor, "Users"); err != nil {
		return nil, err
	}
	v := input.Fields()
	return s.saveUser(ctx, id, v)
}
func (s *Users) Password(ctx context.Context, actor model.Actor, input dto.UserInput, id string) (any, error) {
	if err := s.permission(ctx, actor, "Users"); err != nil {
		return nil, err
	}
	v := input.Fields()
	var err error
	if err = require(v, "new_password"); err != nil {
		return nil, err
	}
	h, err := auth.Hash(model.String(v, "new_password"))
	if err != nil {
		return nil, fail(http.StatusUnprocessableEntity, "密码长度无效")
	}
	_, err = repository.Write(ctx, s.DB, "users", id, model.Record{"hashed_password": h})
	return model.Record{"message": "密码修改成功"}, err
}
func (s *Users) Delete(ctx context.Context, actor model.Actor, id string) (any, error) {
	if err := s.permission(ctx, actor, "Users"); err != nil {
		return nil, err
	}
	if id == actor.ID {
		return nil, fail(http.StatusBadRequest, "不能删除当前用户")
	}
	return model.Record{}, repository.Delete(ctx, s.DB, "users", id)
}
func (s *Users) Import(ctx context.Context, actor model.Actor, upload dto.Upload) (any, error) {
	return s.importUsers(ctx, actor, upload)
}
func (s *Users) Permissions() (any, error) {
	var n navData
	_ = json.Unmarshal(navigation, &n)
	out := []model.Record{}
	for _, m := range n.Menu {
		if model.String(m, "key") != "" && model.String(m, "type") != "divider" {
			out = append(out, model.Record{"key": m["key"], "label": m["label"]})
		}
	}
	return out, nil
}

func (s *Users) serializeUser(ctx context.Context, u model.Record) (model.Record, error) {
	delete(u, "hashed_password")
	roles, err := repository.UserRoleSummary(ctx, s.DB, model.String(u, "id"))
	if err != nil {
		return nil, err
	}
	u["role"], u["role_id"] = nil, nil
	if len(roles) > 0 {
		u["role"] = model.Record{"id": roles[0]["id"], "name": roles[0]["name"]}
		u["role_id"] = roles[0]["id"]
	}
	return u, nil
}

func (s *Users) tokens(u model.Record) (model.Record, error) {
	a, err := auth.Sign(s.Config.Secret, model.String(u, "username"), "access", s.Config.AccessTTL)
	if err != nil {
		return nil, err
	}
	b, err := auth.Sign(s.Config.Secret, model.String(u, "username"), "refresh", s.Config.RefreshTTL)
	return model.Record{"access_token": a, "refresh_token": b, "token_type": "bearer"}, err
}

func (s *Users) saveUser(ctx context.Context, id string, v model.Record) (model.Record, error) {
	if id == "" {
		if err := require(v, "username", "password"); err != nil {
			return nil, err
		}
		if model.String(v, "email") == "" && model.String(v, "phone") == "" {
			return nil, fail(http.StatusUnprocessableEntity, "需要邮箱或手机号码")
		}
		hash, err := auth.Hash(model.String(v, "password"))
		if err != nil {
			return nil, fail(http.StatusUnprocessableEntity, "密码长度无效")
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
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	u, err := repository.Write(ctx, tx, "users", id, v)
	if err != nil {
		return nil, err
	}
	if role, ok := v["role_id"]; ok {
		if _, err = repository.ClearUserRoles(ctx, tx, model.String(u, "id")); err != nil {
			return nil, err
		}
		if role != nil && role != "" {
			if _, err = repository.Write(ctx, tx, "user_roles", "", model.Record{"user_id": u["id"], "role_id": role}); err != nil {
				return nil, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.serializeUser(ctx, u)
}
