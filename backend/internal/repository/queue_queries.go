package repository

import (
	"context"

	"mixm/backend/internal/model"
)

func OutboxTasks(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(ctx, db, "SELECT id FROM instance_tasks WHERE status IN ('pending','processing') AND (published_at IS NULL OR published_at<now()-interval '10 minutes') ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED")
}
func MarkTaskPublished(ctx context.Context, db *DB, taskID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET published_at=now() WHERE id=$1", taskID)
}
func MarkTaskProcessing(ctx context.Context, db *DB, taskID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET status='processing',started_at=COALESCE(started_at,now()),attempts=attempts+1 WHERE id=$1", taskID)
}
func MarkInstanceRunning(ctx context.Context, db *DB, instanceID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instances SET status='running',started_at=COALESCE(started_at,now()),updated_at=now() WHERE id=$1", instanceID)
}
func CompleteTask(ctx context.Context, db *DB, taskID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET status='completed',completed_at=now(),error=NULL WHERE id=$1", taskID)
}
func MarkInstanceFailed(ctx context.Context, db *DB, instanceID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instances SET status='failed',updated_at=now() WHERE id=$1", instanceID)
}
func FailTask(ctx context.Context, db *DB, taskID, message string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET status='failed',completed_at=now(),error=$2 WHERE id=$1", taskID, message)
}
