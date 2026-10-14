package api

import (
	"fmt"
	"net/http"
	"strings"

	"mixm-edu/backend/internal/store"
)

func (s *Server) docsRoutes() {
	s.route("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) (any, error) {
		paths := store.Row{}
		for _, pattern := range s.patterns {
			parts := strings.SplitN(pattern, " ", 2)
			if len(parts) != 2 || !strings.HasPrefix(parts[1], "/api/") {
				continue
			}
			method, path := strings.ToLower(parts[0]), strings.ReplaceAll(parts[1], "...", "")
			operation := store.Row{"responses": store.Row{"200": store.Row{"description": "成功"}, "401": store.Row{"description": "认证失败"}, "422": store.Row{"description": "参数无效"}}, "security": []store.Row{{"bearerAuth": []string{}}}}
			params := []store.Row{}
			for _, segment := range strings.Split(path, "/") {
				if strings.HasPrefix(segment, "{") {
					params = append(params, store.Row{"name": strings.Trim(segment, "{}"), "in": "path", "required": true, "schema": store.Row{"type": "string"}})
				}
			}
			if len(params) > 0 {
				operation["parameters"] = params
			}
			if method == "post" || method == "put" {
				operation["requestBody"] = store.Row{"required": true, "content": store.Row{"application/json": store.Row{"schema": store.Row{"type": "object"}}}}
			}
			if path == "/api/auth/login" || path == "/api/auth/refresh" {
				operation["security"] = []store.Row{}
			}
			item := store.Object(paths[path])
			item[method] = operation
			paths[path] = item
		}
		return store.Row{"openapi": "3.0.3", "info": store.Row{"title": "MutiM Go API", "version": "2.0.0"}, "paths": paths, "components": store.Row{"securitySchemes": store.Row{"bearerAuth": store.Row{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}}}}, nil
	}, true)
	s.mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html lang="zh"><head><meta charset="utf-8"><title>MutiM API</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui',persistAuthorization:true})</script></body></html>`)
	})
}
