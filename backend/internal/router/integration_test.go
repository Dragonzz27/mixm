package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	kubefake "k8s.io/client-go/kubernetes/fake"

	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/infra/cluster"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
	"mixm-edu/backend/internal/service"
	"mixm-edu/backend/internal/testutil"
)

func TestPostgresAPIWorkflow(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	hash, err := auth.Hash("admin-pass")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := repository.Write(ctx, db, "users", "", model.Record{"username": "admin", "email": "admin@example.com", "hashed_password": hash, "is_superuser": true})
	if err != nil {
		t.Fatal(err)
	}
	c := config.Config{Secret: "integration-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour, SSHPort: 30022, SSHHost: "localhost", CodeHost: "localhost", UploadDir: t.TempDir(), ExportDir: t.TempDir()}
	s := New(c, db, nil, &cluster.Client{K: kubefake.NewSimpleClientset()})
	call := func(method, path, token string, payload any, status int) model.Record {
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
		var out model.Record
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return model.Object(out["data"])
	}
	login := call("POST", "/api/auth/login", "", model.Record{"email": "admin@example.com", "password": "admin-pass"}, 200)
	token := model.String(login, "access_token")
	if token == "" {
		t.Fatal("no access token")
	}
	for _, res := range service.ResourcesList {
		call("GET", "/api/"+res.Path, token, nil, 200)
	}
	call("GET", "/api/users", model.String(login, "refresh_token"), nil, 401)
	role := call("POST", "/api/roles", token, model.Record{"name": "students", "route_keys": []string{"ClientHome", "ClientAssignments", "ClientSimulationTest", "ClientMyInstances"}}, 201)
	call("POST", "/api/users", token, model.Record{"username": "student", "email": "student@example.com", "password": "student-pass", "role_id": role["id"]}, 201)
	studentLogin := call("POST", "/api/auth/login", "", model.Record{"email": "student@example.com", "password": "student-pass"}, 200)
	studentToken := model.String(studentLogin, "access_token")
	call("POST", "/api/roles", studentToken, model.Record{"name": "forbidden"}, 403)
	call("GET", "/api/users", studentToken, nil, 403)
	cat := call("POST", "/api/categories", token, model.Record{"name": "Go"}, 201)
	content := call("POST", "/api/contents", token, model.Record{"title": "goroutines", "status": "published"}, 201)
	call("POST", "/api/bindings", token, model.Record{"content_id": content["id"], "category_id": cat["id"]}, 201)
	bindings := call("GET", "/api/binding-counts?category_ids="+model.String(cat, "id"), token, nil, 200)
	if model.Number(bindings, model.String(cat, "id")) != 1 {
		t.Fatal("binding counts incorrect")
	}
	call("PATCH", "/api/categories/"+model.String(cat, "id"), token, model.Record{"parent_id": cat["id"]}, 422)
	assignment := call("POST", "/api/assignments", token, model.Record{"title": "Go homework", "deadline": time.Now().Add(time.Hour).Format(time.RFC3339), "visible_roles": []any{role["id"]}, "attachments": []model.Record{}}, 201)
	submission := call("POST", "/api/assignments/"+model.String(assignment, "id")+"/submissions", studentToken, model.Record{"assignment_id": assignment["id"], "content": "answer", "grade": 100}, 201)
	if submission["grade"] != nil {
		t.Fatal("student assigned own grade")
	}
	review := call("PATCH", "/api/submissions/"+model.String(submission, "id"), token, model.Record{"grade": 88, "feedback": "ok"}, 200)
	if model.Number(review, "grade") != 88 {
		t.Fatal("review grade missing")
	}
	call("POST", "/api/assignments/"+model.String(assignment, "id")+"/submissions", studentToken, model.Record{"assignment_id": assignment["id"], "content": "duplicate"}, 409)
	q := call("POST", "/api/questions", token, model.Record{"title": "选择 Go", "type": "single_choice", "difficulty": 3, "knowledge_tag_ids": []any{cat["id"]}, "options": []model.Record{{"option_key": "A", "content": "Go", "is_correct": true}, {"option_key": "B", "content": "Python", "is_correct": false}}}, 201)
	call("PUT", "/api/questions/"+model.String(q, "id")+"/review", token, model.Record{"review_status": "approved"}, 200)
	paper := call("POST", "/api/simulation-attempts", studentToken, model.Record{"question_count": 1, "course_id": cat["id"]}, 201)
	questions := model.Objects(paper["questions"])
	if len(questions) != 1 {
		t.Fatal("question selection failed")
	}
	resume := call("POST", "/api/simulation-attempts", studentToken, model.Record{"question_count": 1}, 200)
	if resume["attempt_id"] != paper["attempt_id"] {
		t.Fatal("unfinished attempt not resumed")
	}
	tampered := model.Record{"id": q["id"], "correct_index": 1, "score": 10000}
	result := call("POST", "/api/simulation-attempts/"+model.String(paper, "attempt_id")+"/submissions", studentToken, model.Record{"attempt_id": paper["attempt_id"], "questions": []model.Record{tampered}, "answers": []model.Record{{"question_id": q["id"], "selected_index": 1}}}, 200)
	if model.Number(result, "score") != 0 || model.Number(result, "total_score") != 100 {
		t.Fatalf("client controlled grading: %v", result)
	}
	if model.Number(result, "answered_questions") != 1 || model.Number(result, "accuracy") != 0 {
		t.Fatalf("submission statistics incorrect: %v", result)
	}
	accuracy := model.Object(result["knowledge_accuracy"])
	if len(accuracy) != 1 || model.Number(accuracy, "Go") != 0 {
		t.Fatalf("per-test knowledge accuracy lost: %v", accuracy)
	}
	for _, key := range []string{"knowledge_mastery", "q_matrix", "q_matrix_by_question", "q_matrix_by_tag"} {
		if _, exists := result[key]; exists {
			t.Errorf("retired field %s returned in submission", key)
		}
	}
	studentReview := call("GET", "/api/simulation-attempts/"+model.String(paper, "attempt_id")+"/review", studentToken, nil, 200)
	if len(model.Objects(studentReview["questions"])) != 1 || len(model.Objects(studentReview["answers"])) != 1 {
		t.Fatalf("answer review lost: %v", studentReview)
	}
	call("POST", "/api/simulation-attempts/"+model.String(paper, "attempt_id")+"/submissions", studentToken, model.Record{"attempt_id": paper["attempt_id"], "answers": []model.Record{}}, 409)
	call("GET", "/api/simulation-attempts/"+model.String(paper, "attempt_id")+"/review", token, nil, 404)
	call("GET", "/api/questions?difficulty=3", token, nil, 200)
	call("GET", "/api/contents?is_active=true", token, nil, 200)
	task := call("POST", "/api/instances", studentToken, model.Record{"name": "lab", "image": "ubuntu:24.04", "cpu_limit": 2.5}, 202)
	call("GET", "/api/instance-tasks/"+model.String(task, "task_id"), studentToken, nil, 200)
	call("GET", "/api/instance-tasks/"+model.String(task, "task_id"), token, nil, 403)
	stored, err := repository.Get(ctx, db, "instance_tasks", model.String(task, "task_id"))
	if err != nil || stored["published_at"] != nil {
		t.Fatal("unpublished task was not persisted")
	}
	call("POST", "/api/instances", studentToken, model.Record{"user_id": owner["id"]}, 403)
	_, err = db.Exec(ctx, "DELETE FROM contents WHERE id=$1", content["id"])
	if err != nil {
		t.Fatal(err)
	}
	count, err := repository.One(ctx, db, "SELECT count(*) AS total FROM content_categories WHERE content_id=$1", content["id"])
	if err != nil || model.Number(count, "total") != 0 {
		t.Fatal("binding cascade failed", err)
	}
}
