package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestLLMStreamingAdapter(t *testing.T) {
	s := New(config.Config{LLMURL: "http://llm.invalid/v1", LLMKey: "test-key", LLMModel: "test-model"}, nil, nil, nil)
	s.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("LLM request configuration incorrect")
		}
		var payload store.Row
		if e := json.NewDecoder(r.Body).Decode(&payload); e != nil || payload["stream"] != true {
			t.Fatal("stream flag missing")
		}
		event := store.Row{"choices": []store.Row{{"delta": store.Row{"content": `{"title":"test","type":"single_choice","difficulty":3}`}}}}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("data: " + store.JSON(event) + "\n\ndata: [DONE]\n\n")), Header: http.Header{}}, nil
	})
	req := httptest.NewRequest("POST", "/api/questions/generate", strings.NewReader(`{"knowledge_items":[{"name":"Go"}],"difficulty":3}`))
	req = req.WithContext(context.WithValue(req.Context(), userKey{}, store.Row{"is_superuser": true}))
	w := httptest.NewRecorder()
	v, e := s.llmStream(false)(w, req)
	if e != nil || v != nil {
		t.Fatal(e)
	}
	if w.Header().Get("Content-Type") != "text/event-stream" || !strings.Contains(w.Body.String(), `"type":"chunk"`) || !strings.Contains(w.Body.String(), `"type":"done"`) {
		t.Fatal("incompatible SSE response", w.Body.String())
	}
}
func TestRegistryAdapter(t *testing.T) {
	s := New(config.Config{RegistryURL: "http://registry.invalid", RegistryPrefix: "registry.invalid"}, nil, nil, nil)
	s.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body := `{"tags":["latest"]}`
		if strings.Contains(r.URL.Path, "/manifests/") {
			body = `{"config":{"size":10},"layers":[{"size":20},{"size":30}]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})
	tags, e := s.tags(httptest.NewRequest("GET", "/api/images/id/tags", nil), store.Row{"name": "student/lab"})
	if e != nil || len(tags) != 1 || store.Number(tags[0], "size") != 60 {
		t.Fatalf("manifest size incorrect: %v %v", tags, e)
	}
	s.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
	})
	tags, e = s.tags(httptest.NewRequest("GET", "/", nil), store.Row{"name": "new-repo"})
	if e != nil || len(tags) != 0 {
		t.Fatal("empty repository must have no tags")
	}
}
func TestLegacyTemplateNormalization(t *testing.T) {
	v := store.Row{"cpu_limit": "2", "memory_limit_str": "2Gi", "storage": "5Gi", "gpu_limit": "0"}
	if e := normalizeTemplate(v); e != nil {
		t.Fatal(e)
	}
	if e := cluster.Validate(v); e != nil {
		t.Fatal(e)
	}
	if store.Number(v, "storage_limit") != 5 || v["memory_limit"] != "2Gi" {
		t.Fatalf("legacy template not normalized: %v", v)
	}
}
