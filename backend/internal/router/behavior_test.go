package router

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
	"mixm-edu/backend/internal/testutil"
)

func TestCollectionFiltersUpdatesAndBindingTransactions(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	owner, err := repository.Write(ctx, db, "users", "", model.Record{"username": "admin", "hashed_password": "unused", "is_superuser": true})
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.Sign("secret", model.String(owner, "username"), "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	router := New(config.Config{Secret: "secret"}, db, nil, nil)
	call := func(method, path string, payload any, status int) model.Record {
		t.Helper()
		raw, _ := json.Marshal(payload)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != status {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		var body model.Record
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return model.Object(body["data"])
	}
	category := call("POST", "/api/categories/", model.Record{"name": "root"}, 201)
	id := model.String(category, "id")
	call("POST", "/api/categories", model.Record{"name": "child", "parent_id": id}, 201)
	page := call("GET", "/api/categories?offset=1&limit=1", nil, 200)
	if model.Number(page, "total") != 2 || model.Number(page, "offset") != 1 || len(model.Objects(page["items"])) != 1 {
		t.Fatal(page)
	}
	call("GET", "/api/categories?offset=-1", nil, 422)
	updated := call("PATCH", "/api/categories/"+id, model.Record{"is_active": false, "description": nil, "sort_order": 0}, 200)
	if updated["is_active"] != false || updated["description"] != nil || model.Number(updated, "sort_order") != 0 {
		t.Fatal(updated)
	}
	call("PATCH", "/api/categories/"+id, model.Record{"is_active": true}, 200)
	content := call("POST", "/api/contents", model.Record{"title": "lesson", "status": "published"}, 201)
	contentID := model.String(content, "id")
	call("POST", "/api/categories/"+id+"/content-bindings", model.Record{"content_ids": []string{contentID}}, 200)
	detail := call("GET", "/api/categories/"+id+"?include=contents", nil, 200)
	if len(model.Objects(detail["contents"])) != 1 {
		t.Fatal(detail)
	}
	contents := call("GET", "/api/contents?category_ids="+id, nil, 200)
	if len(model.Objects(contents["items"])) != 1 {
		t.Fatal(contents)
	}
	call("POST", "/api/categories/"+id+"/content-bindings", model.Record{"content_ids": []string{contentID, "missing-content"}}, 422)
	call("DELETE", "/api/categories/"+id+"/content-bindings", model.Record{"content_ids": []string{contentID}}, 200)
	unbound := call("GET", "/api/contents?unbound_category_id="+id, nil, 200)
	if len(model.Objects(unbound["items"])) != 1 {
		t.Fatal(unbound)
	}
	call("POST", "/api/assignments", model.Record{"title": "expired", "deadline": time.Now().Add(-time.Hour).Format(time.RFC3339)}, 201)
	call("POST", "/api/assignments", model.Record{"title": "active", "deadline": time.Now().Add(time.Hour).Format(time.RFC3339)}, 201)
	active := call("GET", "/api/assignments?active=true", nil, 200)
	if model.Number(active, "total") != 1 {
		t.Fatal(active)
	}
}

func TestHistoricalFilesAndMultipart(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	user, err := repository.Write(ctx, db, "users", "", model.Record{"username": "files", "hashed_password": "unused"})
	if err != nil {
		t.Fatal(err)
	}
	token, _ := auth.Sign("secret", model.String(user, "username"), "access", time.Hour)
	directory := t.TempDir()
	router := New(config.Config{Secret: "secret", UploadDir: directory, ExportDir: directory}, db, nil, nil)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "lesson.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(part, "attachment bytes")
	form.Close()
	req := httptest.NewRequest("POST", "/api/assignments/files", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result model.Record
	_ = json.Unmarshal(w.Body.Bytes(), &result)
	file := model.Object(result["data"])
	newURL := model.String(file, "file_url")
	name := filepath.Base(newURL)
	for _, path := range []string{newURL, "/api/assignments/file/" + name} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		if recorder.Code != 200 || recorder.Body.String() != "attachment bytes" {
			t.Fatalf("%s: %d %s", path, recorder.Code, recorder.Body.String())
		}
	}

	for _, path := range []string{newURL, "/api/assignments/file/" + name} {
		head := httptest.NewRecorder()
		router.ServeHTTP(head, httptest.NewRequest("HEAD", path, nil))
		if head.Code != 200 || head.Body.Len() != 0 {
			t.Fatal("HEAD file response", head.Code, head.Body.String())
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "legacy.csv"), []byte("csv bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/static/export/legacy.csv", nil))
	if w.Code != 200 || w.Body.String() != "csv bytes" {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestSSEAndOpenAPI(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	user, err := repository.Write(ctx, db, "users", "", model.Record{"username": "llm", "hashed_password": "unused", "is_superuser": true})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input model.Record
		_ = json.NewDecoder(r.Body).Decode(&input)
		if input["stream"] != true {
			t.Error("missing stream flag")
		}
		content := `{"title":"generated"}`
		chunk := model.Record{"choices": []model.Record{{"delta": model.Record{"content": content}}}}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: "+model.JSON(chunk)+"\n\ndata: [DONE]\n\n")
	}))
	defer upstream.Close()
	token, _ := auth.Sign("secret", model.String(user, "username"), "access", time.Hour)
	router := New(config.Config{Secret: "secret", LLMURL: upstream.URL, LLMKey: "test", LLMModel: "test"}, db, nil, nil)
	req := httptest.NewRequest("POST", "/api/questions/generations", strings.NewReader(`{"knowledge_items":[{"name":"Go"}],"question_type":"single_choice","difficulty":3}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 || w.Header().Get("Content-Type") != "text/event-stream" || !strings.Contains(w.Body.String(), `"type":"done"`) || strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatal(w.Code, w.Body.String())
	}
	noLLM := New(config.Config{Secret: "secret"}, db, nil, nil)
	req = httptest.NewRequest("POST", "/api/questions/generations", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	noLLM.ServeHTTP(w, req)
	if w.Code != 503 || !strings.Contains(w.Body.String(), `"code":503`) {
		t.Fatal(w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/openapi.json", nil))
	var doc model.Record
	_ = json.Unmarshal(w.Body.Bytes(), &doc)
	paths := model.Object(doc["paths"])

	for _, path := range []string{"/api/categories/{id}/content-bindings", "/api/simulation-attempts/{id}/submissions"} {
		operation := model.Object(model.Object(paths[path])["post"])
		if model.Object(operation["responses"])["200"] == nil {
			t.Fatalf("wrong POST status documented for %s", path)
		}
	}
	upload := model.Object(model.Object(paths["/api/assignments/files"])["post"])
	if model.Object(model.Object(upload["requestBody"])["content"])["multipart/form-data"] == nil {
		t.Fatal("multipart missing")
	}
	generation := model.Object(model.Object(paths["/api/questions/generations"])["post"])
	if !strings.Contains(model.JSON(generation), "text/event-stream") {
		t.Fatal("SSE documentation missing")
	}
}

func TestCachedTasksRequireOwnership(t *testing.T) {
	if os.Getenv("TEST_REDIS_URL") == "" {
		t.Skip("set TEST_REDIS_URL for task cache integration")
	}
	db := testutil.Database(t)
	ctx := context.Background()
	owner, err := repository.Write(ctx, db, "users", "", model.Record{"username": "owner", "hashed_password": "unused"})
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := repository.Write(ctx, db, "users", "", model.Record{"username": "foreign", "hashed_password": "unused"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := repository.Write(ctx, db, "instance_tasks", "", model.Record{"user_id": owner["id"], "payload": model.Record{"name": "stored"}, "status": "completed"})
	if err != nil {
		t.Fatal(err)
	}
	opts, err := redis.ParseURL(os.Getenv("TEST_REDIS_URL"))
	if err != nil {
		t.Fatal(err)
	}
	cache := redis.NewClient(opts)
	defer cache.Close()
	key := "instance_task:" + model.String(task, "id")
	defer cache.Del(ctx, key)
	cached := model.Record{"id": task["id"], "user_id": owner["id"], "status": "completed", "payload": model.Record{"name": "cached"}}
	if err := cache.Set(ctx, key, model.JSON(cached), time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	router := New(config.Config{Secret: "secret"}, db, cache, nil)
	for _, test := range []struct {
		username string
		status   int
	}{{model.String(owner, "username"), 200}, {model.String(foreign, "username"), 403}} {
		token, _ := auth.Sign("secret", test.username, "access", time.Hour)
		req := httptest.NewRequest("GET", "/api/instance-tasks/"+model.String(task, "id"), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != test.status {
			t.Fatal(test.username, w.Code, w.Body.String())
		}
		if w.Code == 200 && !strings.Contains(w.Body.String(), "cached") {
			t.Fatal("owner cache miss")
		}
	}
}
