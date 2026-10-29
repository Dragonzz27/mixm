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

func TestAPIRouteCoverage(t *testing.T) {
	router := New(config.Config{Secret: "test"}, nil, nil, nil)
	registered := map[string]bool{}
	for _, route := range router.Engine.Routes() {
		registered[route.Method+" "+route.Path] = true
	}
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
func TestHTTPContract(t *testing.T) {
	router := New(config.Config{Secret: "test"}, nil, nil, nil)
	for _, path := range []string{"/api/users", "/api/categories/", "/api/instance-tasks", "/api/simulation-attempts/current", "/api/clusters/nodes", "/api/clusters/persistent-volume-claims"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 || !strings.Contains(w.Body.String(), `"code":401`) || strings.Contains(w.Body.String(), `"detail"`) {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
	for _, test := range []struct {
		method, path string
		status       int
	}{{"OPTIONS", "/api/categories", 204}, {"GET", "/api/auth/users", 404}, {"POST", "/api/instances/start", 405}, {"GET", "/missing", 404}, {"GET", "/api/auth/login", 405}, {"GET", "/api/knowledge-tracing/status", 404}, {"GET", "/api/simulation-traces", 404}} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(test.method, test.path, nil))
		if w.Code != test.status {
			t.Errorf("%s %s: %d", test.method, test.path, w.Code)
		}
	}
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
