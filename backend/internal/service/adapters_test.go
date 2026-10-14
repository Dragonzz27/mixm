package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/dto"
	"mixm-edu/backend/internal/infra/cluster"
	"mixm-edu/backend/internal/model"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestLLMStreamingAdapter(t *testing.T) {
	s := NewCore(config.Config{LLMURL: "http://llm.invalid/v1", LLMKey: "test-key", LLMModel: "test-model"}, nil, nil, nil)
	s.Clients.HTTP.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("LLM request configuration incorrect")
		}
		var payload model.Record
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload["stream"] != true {
			t.Fatal("stream flag missing")
		}
		event := model.Record{"choices": []model.Record{{"delta": model.Record{"content": `{"title":"test","type":"single_choice","difficulty":3}`}}}}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("data: " + model.JSON(event) + "\n\ndata: [DONE]\n\n")), Header: http.Header{}}, nil
	})
	var input dto.LearningInput
	if err := json.Unmarshal([]byte(`{"knowledge_items":[{"name":"Go"}],"difficulty":3}`), &input); err != nil {
		t.Fatal(err)
	}
	events := []model.Record{}
	learning := &Learning{Core: s}
	value, err := learning.QuestionGeneration(context.Background(), model.Actor{IsSuperuser: true}, input, func(event model.Record) { events = append(events, event) })
	if err != nil || value != nil || len(events) != 2 || events[0]["type"] != "chunk" || events[1]["type"] != "done" {
		t.Fatalf("invalid stream: %v %v", events, err)
	}

}
func TestRegistryAdapter(t *testing.T) {
	s := NewCore(config.Config{RegistryURL: "http://registry.invalid", RegistryPrefix: "registry.invalid"}, nil, nil, nil)
	s.Clients.HTTP.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body := `{"tags":["latest"]}`
		if strings.Contains(r.URL.Path, "/manifests/") {
			body = `{"config":{"size":10},"layers":[{"size":20},{"size":30}]}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})
	tags, err := (&Images{Core: s}).tags(context.Background(), model.Record{"name": "student/lab"})
	if err != nil || len(tags) != 1 || model.Number(tags[0], "size") != 60 {
		t.Fatalf("manifest size incorrect: %v %v", tags, err)
	}
	s.Clients.HTTP.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
	})
	tags, err = (&Images{Core: s}).tags(context.Background(), model.Record{"name": "new-repo"})
	if err != nil || len(tags) != 0 {
		t.Fatal("empty repository must have no tags")
	}
}
func TestTemplateNormalization(t *testing.T) {
	v := model.Record{"cpu_limit": "2", "memory_limit_str": "2Gi", "storage": "5Gi", "gpu_limit": "0"}
	if err := normalizeTemplate(v); err != nil {
		t.Fatal(err)
	}
	if err := cluster.Validate(v); err != nil {
		t.Fatal(err)
	}
	if model.Number(v, "storage_limit") != 5 || v["memory_limit"] != "2Gi" {
		t.Fatalf("template configuration not normalized: %v", v)
	}
}
