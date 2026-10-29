package router

import (
	_ "embed"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"mixm/backend/internal/config"
)

//go:embed routes.json
var requiredRoutes []byte

// TestAPIRouteCoverage 验证接口契约中的路由均已注册。
func TestAPIRouteCoverage(t *testing.T) {
	router := New(config.Config{Secret: "test"}, nil, nil, nil)
	registered := map[string]bool{}

	for _, route := range router.Engine.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	// 将实际注册的接口与版本化路由契约逐项比对。
	var expected []string
	if err := json.Unmarshal(requiredRoutes, &expected); err != nil {
		t.Fatal(err)
	}

	for _, route := range expected {
		if !registered[route] {
			t.Errorf("missing endpoint %s", route)
		}
	}
}

// TestHTTPContract 验证认证、请求方法和参数错误的 HTTP 响应。
func TestHTTPContract(t *testing.T) {
	router := New(config.Config{Secret: "test"}, nil, nil, nil)

	// 受保护接口应先返回认证错误，避免路由缺失被误判为空数据。
	for _, path := range []string{
		"/api/users",
		"/api/categories/",
		"/api/instance-tasks",
		"/api/simulation-attempts/current",
		"/api/clusters/nodes",
		"/api/clusters/persistent-volume-claims",
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 || !strings.Contains(w.Body.String(), `"code":401`) || strings.Contains(w.Body.String(), `"detail"`) {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/api/images/sync", nil))
	if w.Code != 401 {
		t.Fatalf("image sync must require authentication: %d", w.Code)
	}

	// 检查预检、请求方法和不存在资源的响应状态。
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{"OPTIONS", "/api/categories", 204},
		{"GET", "/api/auth/users", 404},
		{"POST", "/api/instances/start", 405},
		{"GET", "/missing", 404},
		{"GET", "/api/auth/login", 405},
		{"GET", "/api/knowledge-tracing/status", 404},
		{"GET", "/api/simulation-traces", 404},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(test.method, test.path, nil))
		if w.Code != test.status {
			t.Errorf("%s %s: %d", test.method, test.path, w.Code)
		}
	}

	// 检查无效 JSON 和字段类型错误的参数响应。
	for _, body := range []string{`null`, `[]`, `{}`, `{"email":1}`, `{} {}`} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body)))
		if body == `{}` {
			continue
		}

		if w.Code != 422 {
			t.Errorf("body %s: %d %s", body, w.Code, w.Body.String())
		}
	}
}
