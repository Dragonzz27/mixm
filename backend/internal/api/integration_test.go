package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	kubefake "k8s.io/client-go/kubernetes/fake"
	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
	"mixm-edu/backend/internal/testutil"
)

func TestPostgresAPIWorkflow(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	hash, e := auth.Hash("admin-pass")
	if e != nil {
		t.Fatal(e)
	}
	owner, e := store.Write(ctx, db, "users", "", store.Row{"username": "admin", "email": "admin@example.com", "hashed_password": hash, "is_superuser": true})
	if e != nil {
		t.Fatal(e)
	}
	c := config.Config{Secret: "integration-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour, SSHPort: 30022, SSHHost: "localhost", CodeHost: "localhost", UploadDir: t.TempDir(), ExportDir: t.TempDir()}
	s := New(c, db, nil, &cluster.Client{K: kubefake.NewSimpleClientset()})
	call := func(method, path, token string, payload any, status int) store.Row {
		t.Helper()
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		if w.Code != status {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		var out store.Row
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	login := call("POST", "/api/auth/login", "", store.Row{"email": "admin@example.com", "password": "admin-pass"}, 200)
	token := store.String(login, "access_token")
	if token == "" {
		t.Fatal("no access token")
	}
	for _, res := range resources {
		call("GET", "/api/"+res.Path, token, nil, 200)
	}
	call("GET", "/api/auth/users", store.String(login, "refresh_token"), nil, 401)
	role := call("POST", "/api/roles", token, store.Row{"name": "students", "route_keys": []string{"ClientHome", "ClientAssignments", "ClientSimulationTest", "ClientMyInstances"}}, 201)
	call("POST", "/api/auth/register", token, store.Row{"username": "student", "email": "student@example.com", "password": "student-pass", "role_id": role["id"]}, 200)
	studentLogin := call("POST", "/api/auth/login", "", store.Row{"email": "student@example.com", "password": "student-pass"}, 200)
	studentToken := store.String(studentLogin, "access_token")
	call("POST", "/api/roles", studentToken, store.Row{"name": "forbidden"}, 403)
	call("GET", "/api/auth/users", studentToken, nil, 403)
	cat := call("POST", "/api/categories", token, store.Row{"name": "Go"}, 201)
	content := call("POST", "/api/contents", token, store.Row{"title": "goroutines", "status": "published"}, 201)
	call("POST", "/api/bindings", token, store.Row{"content_id": content["id"], "category_id": cat["id"]}, 200)
	bindings := call("POST", "/api/bindings/counts", token, store.Row{"category_ids": []any{cat["id"]}}, 200)
	if store.Number(bindings, store.String(cat, "id")) != 1 {
		t.Fatal("binding counts incorrect")
	}
	call("PUT", "/api/categories/"+store.String(cat, "id"), token, store.Row{"parent_id": cat["id"]}, 422)
	assignment := call("POST", "/api/assignments", token, store.Row{"title": "Go homework", "deadline": time.Now().Add(time.Hour).Format(time.RFC3339), "visible_roles": []any{role["id"]}, "attachments": []store.Row{}}, 201)
	submission := call("POST", "/api/assignments/submissions", studentToken, store.Row{"assignment_id": assignment["id"], "content": "answer", "grade": 100}, 201)
	if submission["grade"] != nil {
		t.Fatal("student assigned own grade")
	}
	review := call("PUT", "/api/assignments/submissions/"+store.String(submission, "id"), token, store.Row{"grade": 88, "feedback": "ok"}, 200)
	if store.Number(review, "grade") != 88 {
		t.Fatal("review grade missing")
	}
	call("POST", "/api/assignments/submissions", studentToken, store.Row{"assignment_id": assignment["id"], "content": "duplicate"}, 409)
	q := call("POST", "/api/questions", token, store.Row{"title": "选择 Go", "type": "single_choice", "difficulty": 3, "knowledge_tag_ids": []any{cat["id"]}, "options": []store.Row{{"option_key": "A", "content": "Go", "is_correct": true}, {"option_key": "B", "content": "Python", "is_correct": false}}}, 201)
	call("POST", "/api/questions/"+store.String(q, "id")+"/review", token, store.Row{"review_status": "approved"}, 200)
	paper := call("POST", "/api/simulation/generate", studentToken, store.Row{"question_count": 1, "course_id": cat["id"]}, 200)
	questions := store.Objects(paper["questions"])
	if len(questions) != 1 {
		t.Fatal("question selection failed")
	}
	resume := call("POST", "/api/simulation/generate", studentToken, store.Row{"question_count": 1}, 200)
	if resume["attempt_id"] != paper["attempt_id"] {
		t.Fatal("unfinished attempt not resumed")
	}
	tampered := store.Row{"id": q["id"], "correct_index": 1, "score": 10000}
	result := call("POST", "/api/simulation/submit", studentToken, store.Row{"attempt_id": paper["attempt_id"], "questions": []store.Row{tampered}, "answers": []store.Row{{"question_id": q["id"], "selected_index": 1}}}, 200)
	if store.Number(result, "score") != 0 || store.Number(result, "total_score") != 100 {
		t.Fatalf("client controlled grading: %v", result)
	}
	call("POST", "/api/simulation/submit", studentToken, store.Row{"attempt_id": paper["attempt_id"], "answers": []store.Row{}}, 409)
	call("GET", "/api/simulation/attempt/"+store.String(paper, "attempt_id")+"/review", token, nil, 404)
	call("GET", "/api/questions?difficulty=3", token, nil, 200)
	call("GET", "/api/contents?is_active=true", token, nil, 200)
	task := call("POST", "/api/instances/start", studentToken, store.Row{"name": "lab", "image": "ubuntu:24.04", "cpu_limit": 2.5}, 202)
	call("GET", "/api/instances/queue/task/"+store.String(task, "task_id"), studentToken, nil, 200)
	call("GET", "/api/instances/queue/task/"+store.String(task, "task_id"), token, nil, 403)
	stored, e := store.Get(ctx, db, "instance_tasks", store.String(task, "task_id"))
	if e != nil || stored["published_at"] != nil {
		t.Fatal("unpublished task was not persisted")
	}
	call("POST", "/api/instances/start", studentToken, store.Row{"user_id": owner["id"]}, 403)
	_, e = db.Exec(ctx, "DELETE FROM contents WHERE id=$1", content["id"])
	if e != nil {
		t.Fatal(e)
	}
	var count int
	if e = db.QueryRow(ctx, "SELECT count(*) FROM content_categories WHERE content_id=$1", content["id"]).Scan(&count); e != nil || count != 0 {
		t.Fatal("binding cascade failed")
	}
}
