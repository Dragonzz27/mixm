package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mixm/backend/internal/auth"
	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/router"
)

// TestSeedExamplesStudentFlow 验证示例幂等、学生可见、作业可交，并能直接用示例练习组卷答题。
func TestSeedExamplesStudentFlow(t *testing.T) {
	db := commandTestDatabase(t)
	ctx := context.Background()
	for range 2 {
		if err := Seed(); err != nil {
			t.Fatal(err)
		}
	}

	for table, expected := range map[string]int{
		"categories":         4,
		"contents":           3,
		"content_categories": 3,
		"assignments":        3,
		"questions":          5,
		"question_options":   8,
		"question_tags":      5,
		"practices":          1,
		"practice_questions": 5,
		"users":              0,
	} {
		row, err := repository.One(ctx, db, "SELECT count(*) AS total FROM "+table)
		if err != nil || int(model.Number(row, "total")) != expected {
			t.Fatalf("unexpected seed count for %s: %v, %v", table, row, err)
		}
	}

	student, err := repository.Write(ctx, db, "users", "", model.Record{
		"username":        "demo-student",
		"hashed_password": "unused",
	})
	if err != nil {
		t.Fatal(err)
	}
	role, err := repository.One(ctx, db, "SELECT id FROM roles WHERE name=$1", "学生")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Write(ctx, db, "user_roles", "", model.Record{
		"user_id": student["id"],
		"role_id": role["id"],
	}); err != nil {
		t.Fatal(err)
	}

	settings := config.Config{Secret: "test-secret"}
	api := router.New(settings, db, nil, nil)
	token, err := auth.Sign(settings.Secret, model.String(student, "username"), "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path string, payload any, status int) model.Record {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		api.ServeHTTP(response, req)
		if response.Code != status {
			t.Fatalf("%s %s returned %d, want %d: %s", method, path, response.Code, status, response.Body.String())
		}
		var envelope model.Record
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return model.Object(envelope["data"])
	}

	// 学生看到的是已发布正文以及完整的分类关联，而不是管理端专用草稿。
	categories := call("GET", "/api/categories?root_only=true&include=content-preview", nil, 200)
	root := model.Objects(categories["items"])[0]
	preview := call("GET", "/api/categories/"+model.String(root, "id")+"/content-tree", nil, 200)
	if !strings.Contains(model.JSON(preview), "Linux 基础") {
		t.Fatal("seeded categories are missing from the student content tree")
	}
	children := model.Objects(preview["children"])
	if len(children) != 3 {
		t.Fatal("example course does not contain all three lesson categories")
	}
	for _, category := range children {
		bound := call("GET", "/api/categories/"+model.String(category, "id")+"/contents", nil, 200)
		if len(model.Objects(bound["items"])) != 1 {
			t.Fatal("published lesson is missing from its category")
		}
	}
	contents := call("GET", "/api/contents?status=published&is_active=true", nil, 200)
	if model.Number(contents, "total") != 3 {
		t.Fatal("published example lessons are not visible")
	}
	for _, content := range model.Objects(contents["items"]) {
		detail := call("GET", "/api/contents/"+model.String(content, "id"), nil, 200)
		if !strings.HasPrefix(model.String(detail, "content"), "[md]") || len(model.Objects(detail["categories"])) != 1 {
			t.Fatal("lesson body or category binding is missing")
		}
	}

	assignments := call("GET", "/api/assignments?active=true", nil, 200)
	if model.Number(assignments, "total") != 3 {
		t.Fatal("example assignments are missing or already expired")
	}
	assignmentID := model.String(model.Objects(assignments["items"])[0], "id")
	call("POST", "/api/assignments/"+assignmentID+"/submissions", model.Record{
		"content": "已完成示例操作并记录执行结果。",
	}, 201)

	practices := call("GET", "/api/practices", nil, 200)
	practice := model.Objects(practices["items"])[0]
	if model.Number(practice, "question_count") != 5 {
		t.Fatal("example practice does not contain all five question types")
	}
	paper := call("POST", "/api/simulation-attempts", model.Record{
		"generation_mode": "practice",
		"practice_id":     practice["id"],
		"question_count":  10,
	}, 201)
	questions := model.Objects(paper["questions"])
	if len(questions) != 5 {
		t.Fatal("example practice could not generate a complete paper")
	}
	answers := []model.Record{}
	for _, question := range questions {
		answer := model.Record{"question_id": question["id"]}
		switch model.String(question, "question_type") {
		case "multi_choice":
			answer["answer_text"] = model.JSON(question["correct_indices"])
		case "fill_blank", "short_answer":
			answer["answer_text"] = question["correct_answer"]
		default:
			answer["selected_index"] = question["correct_index"]
		}
		answers = append(answers, answer)
	}
	result := call("POST", "/api/simulation-attempts/"+model.String(paper, "attempt_id")+"/submissions", model.Record{
		"answers": answers,
	}, 200)
	if model.Number(result, "score") != 100 || model.Number(result, "correct_count") != 5 {
		t.Fatal("seeded question answers or options cannot be scored correctly")
	}

	// 再次初始化保留管理员修改的标题、截止时间及题目选项。
	content := model.Objects(contents["items"])[0]
	if _, err := repository.Write(ctx, db, "contents", model.String(content, "id"), model.Record{
		"title":  "管理员已编辑的教程",
		"status": "draft",
	}); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	if _, err := repository.Write(ctx, db, "assignments", assignmentID, model.Record{
		"deadline": past,
		"status":   "closed",
	}); err != nil {
		t.Fatal(err)
	}
	options, err := repository.Query(ctx, db, "SELECT id FROM question_options ORDER BY id LIMIT 1")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Delete(ctx, db, "question_options", model.String(options[0], "id")); err != nil {
		t.Fatal(err)
	}
	if err := Seed(); err != nil {
		t.Fatal(err)
	}
	edited, err := repository.Get(ctx, db, "contents", model.String(content, "id"))
	if err != nil || edited["title"] != "管理员已编辑的教程" || edited["status"] != "draft" {
		t.Fatal("reseed overwrote an edited lesson")
	}
	saved, err := repository.Get(ctx, db, "assignments", assignmentID)
	if err != nil || saved["status"] != "closed" || !saved["deadline"].(time.Time).Equal(past) {
		t.Fatal("reseed changed assignment status or deadline")
	}
	count, err := repository.One(ctx, db, "SELECT count(*) AS total FROM question_options")
	if err != nil || model.Number(count, "total") != 7 {
		t.Fatal("reseed restored a deliberately removed question option")
	}
}

// TestSeedExamplesRollback 验证示例写入失败时，角色、分类和正文也一并回滚。
func TestSeedExamplesRollback(t *testing.T) {
	db := commandTestDatabase(t)
	ctx := context.Background()
	if _, err := db.Exec(ctx, `CREATE FUNCTION reject_example_question() RETURNS trigger AS $$
BEGIN RAISE EXCEPTION 'reject example question'; END;
$$ LANGUAGE plpgsql`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `CREATE TRIGGER reject_example_question BEFORE INSERT ON questions
FOR EACH ROW EXECUTE FUNCTION reject_example_question()`); err != nil {
		t.Fatal(err)
	}
	if err := Seed(); err == nil {
		t.Fatal("seed must fail when a related example cannot be inserted")
	}
	for _, table := range []string{"roles", "content_types", "categories", "contents", "assignments", "questions"} {
		row, err := repository.One(ctx, db, "SELECT count(*) AS total FROM "+table)
		if err != nil || model.Number(row, "total") != 0 {
			t.Fatalf("failed seed left partial records in %s", table)
		}
	}
}
