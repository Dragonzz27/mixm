package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (r *Router) OpenAPI(c *gin.Context) {
	paths := gin.H{}
	for _, route := range r.Engine.Routes() {
		if !strings.HasPrefix(route.Path, "/api/") || strings.Contains(route.Path, "/assignments/file/") || strings.Contains(route.Path, "/contents/file/") {
			continue
		}
		path := route.Path
		params := []gin.H{}
		for _, segment := range strings.Split(path, "/") {
			if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
				name := segment[1:]
				path = strings.Replace(path, segment, "{"+name+"}", 1)
				params = append(params, gin.H{"name": name, "in": "path", "required": true, "schema": gin.H{"type": "string"}})
			}
		}
		streaming := strings.HasSuffix(path, "/questions/generations") || strings.HasSuffix(path, "/knowledge-tree/analyses")
		file := (route.Method == "GET" || route.Method == "HEAD") && (strings.Contains(path, "/files/") || strings.Contains(path, "/file/"))
		media := "application/json"
		if streaming {
			media = "text/event-stream"
		}
		if file {
			media = "application/octet-stream"
		}
		status := "200"
		if route.Method == "POST" && !streaming && !strings.HasPrefix(path, "/api/auth/") {
			status = "201"
		}
		if route.Method == "POST" && (path == "/api/instances" || strings.HasSuffix(path, "/instances") && strings.HasPrefix(path, "/api/templates/")) {
			status = "202"
		}

		if route.Method == "POST" && (strings.HasSuffix(path, "/content-bindings") || strings.HasPrefix(path, "/api/simulation-attempts/") && strings.HasSuffix(path, "/submissions")) {
			status = "200"
		}
		schema := gin.H{"$ref": "#/components/schemas/Envelope"}
		if file {
			schema = gin.H{"type": "string", "format": "binary"}
		}
		if streaming {
			schema = gin.H{"type": "string"}
		}
		responses := gin.H{status: gin.H{"description": "成功", "content": gin.H{media: gin.H{"schema": schema}}}}
		for _, code := range []string{"401", "403", "404", "409", "422", "500", "503"} {
			responses[code] = gin.H{"description": "请求失败", "content": gin.H{"application/json": gin.H{"schema": gin.H{"$ref": "#/components/schemas/Envelope"}}}}
		}
		if route.Method == "POST" && path == "/api/simulation-attempts" {
			responses["200"] = responses["201"]
		}
		operation := gin.H{"responses": responses, "security": []gin.H{{"bearerAuth": []string{}}}}
		list := false
		for _, name := range []string{".List", ".Mine", ".ForUser", ".Tasks", ".Submissions", ".CategoryContents", ".Tags", ".GetPods", ".GetNodes", ".GetDeployments", ".GetPv", ".GetPvc"} {
			if strings.HasSuffix(route.Handler, name+"-fm") {
				list = true
				break
			}
		}
		if route.Method == http.MethodGet && list {
			params = append(params, gin.H{"name": "offset", "in": "query", "schema": gin.H{"type": "integer", "minimum": 0, "default": 0}}, gin.H{"name": "limit", "in": "query", "schema": gin.H{"type": "integer", "minimum": 1, "maximum": 1000, "default": 100}})
		}
		if len(params) > 0 {
			operation["parameters"] = params
		}
		public := strings.HasPrefix(path, "/api/auth/") || file
		if public {
			operation["security"] = []gin.H{}
		}
		if route.Method == "POST" || route.Method == "PATCH" || route.Method == "PUT" || route.Method == "DELETE" && strings.HasSuffix(path, "/content-bindings") {
			requestMedia := "application/json"
			requestSchema := gin.H{"type": "object"}
			if strings.HasSuffix(path, "/files") || strings.HasSuffix(path, "/imports") {
				requestMedia = "multipart/form-data"
				requestSchema = gin.H{"type": "object", "required": []string{"file"}, "properties": gin.H{"file": gin.H{"type": "string", "format": "binary"}}}
			}
			if strings.HasSuffix(path, "/questions/batches") {
				requestSchema = gin.H{"type": "array", "items": gin.H{"type": "object"}}
			}
			operation["requestBody"] = gin.H{"required": true, "content": gin.H{requestMedia: gin.H{"schema": requestSchema}}}
		}
		item, _ := paths[path].(gin.H)
		if item == nil {
			item = gin.H{}
		}
		item[strings.ToLower(route.Method)] = operation
		paths[path] = item
	}
	c.JSON(http.StatusOK, gin.H{"openapi": "3.0.3", "info": gin.H{"title": "mixm Gin API", "version": "3.0.0"}, "paths": paths, "components": gin.H{"securitySchemes": gin.H{"bearerAuth": gin.H{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}}, "schemas": gin.H{"Envelope": gin.H{"type": "object", "required": []string{"code", "message", "data"}, "properties": gin.H{"code": gin.H{"type": "integer"}, "message": gin.H{"type": "string"}, "data": gin.H{"nullable": true}}}}}})
}
func (r *Router) Docs(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html><html lang="zh"><head><meta charset="utf-8"><title>mixm API</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui',persistAuthorization:true})</script></body></html>`))
}
