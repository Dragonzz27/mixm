package repository

import (
	"context"

	"mixm/backend/internal/model"
)

func AllInstances(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM instances")
}
func TaskSummary(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT status,count(*) AS n FROM instance_tasks WHERE user_id=$1 GROUP BY status", userID)
}
