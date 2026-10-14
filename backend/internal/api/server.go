package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

type Server struct {
	Config    config.Config
	DB        *pgxpool.Pool
	Redis     *redis.Client
	Cluster   *cluster.Client
	Transport http.RoundTripper
	mux       *http.ServeMux
	files     *http.ServeMux
	patterns  []string
}
type handler func(http.ResponseWriter, *http.Request) (any, error)
type userKey struct{}
type Error struct {
	Status int
	Detail string
}

func (e *Error) Error() string            { return e.Detail }
func fail(code int, message string) error { return &Error{code, message} }
func user(r *http.Request) store.Row      { u, _ := r.Context().Value(userKey{}).(store.Row); return u }
func admin(r *http.Request) error {
	if !store.Bool(user(r), "is_superuser") {
		return fail(403, "需要管理员权限")
	}
	return nil
}
func body(r *http.Request) (store.Row, error) {
	var v store.Row
	e := decode(r, &v)
	if e != nil {
		return nil, e
	}
	if v == nil {
		return nil, fail(422, "需要 JSON 对象")
	}
	return v, nil
}
func decode(r *http.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(r.Body, 8<<20))
	d.UseNumber()
	if e := d.Decode(v); e != nil {
		return fail(422, "无效的 JSON 请求")
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return fail(422, "请求只能包含一个 JSON 值")
	}
	return nil
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status != 204 {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func (s *Server) route(pattern string, h handler, public ...bool) {
	s.patterns = append(s.patterns, pattern)
	mux := s.mux
	if strings.Contains(pattern, "/api/assignments/file/") {
		mux = s.files
	}
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if len(public) == 0 || !public[0] {
			c, e := auth.Parse(s.Config.Secret, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), "access")
			if e != nil {
				respond(w, 401, store.Row{"detail": "登录已过期"})
				return
			}
			u, e := store.One(r.Context(), s.DB, "SELECT * FROM users WHERE username=$1 AND is_active=true", c.Subject)
			if e != nil {
				respond(w, 401, store.Row{"detail": "用户不可用"})
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), userKey{}, u))
		}
		v, e := h(w, r)
		if e != nil {
			code, detail := 500, "服务器错误"
			var he *Error
			var pe *pgconn.PgError
			switch {
			case errors.As(e, &he):
				code, detail = he.Status, he.Detail
			case errors.Is(e, pgx.ErrNoRows):
				code, detail = 404, "资源不存在"
			case errors.Is(e, store.ErrInvalidValue):
				code, detail = 422, e.Error()
			case errors.As(e, &pe):
				switch pe.Code {
				case "23505":
					code, detail = 409, "数据已存在"
				case "23503", "23514", "23502", "22P02":
					code, detail = 422, "数据或关联资源无效"
				}
			}
			if code == 500 {
				slog.Error("request failed", "path", r.URL.Path, "error", e)
			}
			respond(w, code, store.Row{"detail": detail})
			return
		}
		if v == nil {
			return
		}
		code := 200
		if r.Method == http.MethodDelete {
			code = 204
		}
		respond(w, code, v)
	})
}
func New(c config.Config, db *pgxpool.Pool, cache *redis.Client, k *cluster.Client) *Server {
	s := &Server{Config: c, DB: db, Redis: cache, Cluster: k, mux: http.NewServeMux(), files: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			slog.Error("handler panic", "error", e)
			respond(w, 500, store.Row{"detail": "服务器错误"})
		}
	}()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(204)
		return
	}
	if len(r.URL.Path) > 1 {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
	}
	if strings.HasPrefix(r.URL.Path, "/api/assignments/file/") {
		s.files.ServeHTTP(w, r)
		return
	}
	s.mux.ServeHTTP(w, r)
}
func (s *Server) routes() {
	s.route("GET /health", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.DB.Ping(r.Context()); e != nil {
			return nil, fail(503, "数据库不可用")
		}
		return store.Row{"status": "ok"}, nil
	}, true)
	s.authRoutes()
	s.resourceRoutes()
	s.bindingRoutes()
	s.assignmentRoutes()
	s.instanceRoutes()
	s.clusterRoutes()
	s.learningRoutes()
	s.fileRoutes()
	s.imageRoutes()
	s.systemRoutes()
	s.docsRoutes()
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.mux.Handle("GET /static/export/", http.StripPrefix("/static/export/", http.FileServer(http.Dir(s.Config.ExportDir))))
}
func require(v store.Row, fields ...string) error {
	for _, f := range fields {
		if strings.TrimSpace(store.String(v, f)) == "" {
			return fail(422, fmt.Sprintf("%s 不能为空", f))
		}
	}
	return nil
}

func (s *Server) do(req *http.Request, timeout time.Duration) (*http.Response, error) {
	return (&http.Client{Transport: s.Transport, Timeout: timeout}).Do(req)
}
