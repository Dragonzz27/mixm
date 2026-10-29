package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// AllInstances 读取全部实例记录。
func AllInstances(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM instances")
}

// TaskSummary 按任务状态统计用户的实例创建任务。
func TaskSummary(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT status,count(*) AS n FROM instance_tasks WHERE user_id=$1 GROUP BY status", userID)
}
