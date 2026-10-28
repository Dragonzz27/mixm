package repository_test

import (
	"context"
	"testing"

	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/testutil"
)

func TestRemoveKnowledgeTracingMigrationPreservesTests(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()

	// Recreate retired storage and mark its cleanup pending to simulate an upgrade.
	_, err := db.Exec(ctx, `
		CREATE TABLE knowledge_mastery (id text PRIMARY KEY, mastery double precision);
		INSERT INTO knowledge_mastery VALUES ('old-mastery', 0.7);
		CREATE TABLE simulation_answer_trace (id text PRIMARY KEY, is_correct boolean);
		INSERT INTO simulation_answer_trace VALUES ('old-trace', true);
		DELETE FROM schema_migrations WHERE name='003_remove_knowledge_tracing.sql';
		INSERT INTO simulation_attempt (id, user_id, status, questions_payload, answers_payload, result_payload)
		VALUES ('submitted', 'student', 'submitted', '[{"id":"question"}]',
		  '[{"question_id":"question","selected_index":0}]',
		  '{"score":100,"accuracy":100,"knowledge_accuracy":{"Go":100},"knowledge_mastery":{"Go":0.7},"q_matrix":[{}],"q_matrix_by_question":{},"q_matrix_by_tag":{}}'),
		  ('current', 'student', 'in_progress', '[]', '[]', NULL),
		  ('plain', 'other-student', 'submitted', '[]', '[]', '{"score":50}');
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"knowledge_mastery", "simulation_answer_trace"} {
		row, err := repository.One(ctx, db, "SELECT to_regclass(current_schema() || '.' || $1)::text AS name", table)
		if err != nil || row["name"] != nil {
			t.Fatalf("retired table %s remains: %v, %v", table, row, err)
		}
	}
	row, err := repository.Get(ctx, db, "simulation_attempt", "submitted")
	if err != nil {
		t.Fatal(err)
	}
	result := model.Object(row["result_payload"])
	if len(result) != 3 || model.Number(result, "score") != 100 || model.Number(result, "accuracy") != 100 || model.Number(model.Object(result["knowledge_accuracy"]), "Go") != 100 {
		t.Fatalf("test results changed or retired fields remain: %v", result)
	}
	if len(model.Objects(row["questions_payload"])) != 1 || len(model.Objects(row["answers_payload"])) != 1 || row["status"] != "submitted" {
		t.Fatalf("test history changed: %v", row)
	}
	current, err := repository.Get(ctx, db, "simulation_attempt", "current")
	if err != nil || current["result_payload"] != nil || current["status"] != "in_progress" {
		t.Fatalf("unfinished test changed: %v, %v", current, err)
	}
	plain, err := repository.Get(ctx, db, "simulation_attempt", "plain")
	if err != nil || model.Number(model.Object(plain["result_payload"]), "score") != 50 {
		t.Fatalf("ordinary result changed: %v, %v", plain, err)
	}
	if err = repository.Migrate(ctx, db); err != nil {
		t.Fatalf("repeated startup migration failed: %v", err)
	}
}
