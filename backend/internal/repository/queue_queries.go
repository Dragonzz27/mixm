package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// OutboxTasks 锁定待发布或需重新投递的实例任务。
func OutboxTasks(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT id FROM instance_tasks WHERE status IN ('pending','processing') AND (published_at IS NULL OR published_at<now()-interval '10 minutes') ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED",
	)
}

// MarkTaskPublished 记录任务成功发布到队列的时间。
func MarkTaskPublished(ctx context.Context, db *DB, taskID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET published_at=now() WHERE id=$1", taskID)
}

// MarkTaskProcessing 将任务标记为处理中并累计尝试次数。
func MarkTaskProcessing(ctx context.Context, db *DB, taskID string) (Result, error) {
	return db.Exec(
		ctx,
		"UPDATE instance_tasks SET status='processing',started_at=COALESCE(started_at,now()),attempts=attempts+1 WHERE id=$1",
		taskID,
	)
}

// MarkInstanceRunning 将实例标记为运行中并记录启动时间。
func MarkInstanceRunning(ctx context.Context, db *DB, instanceID string) (Result, error) {
	return db.Exec(
		ctx,
		"UPDATE instances SET status='running',started_at=COALESCE(started_at,now()),updated_at=now() WHERE id=$1",
		instanceID,
	)
}

// CompleteTask 记录任务完成时间并清除错误信息。
func CompleteTask(ctx context.Context, db *DB, taskID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET status='completed',completed_at=now(),error=NULL WHERE id=$1", taskID)
}

// MarkInstanceFailed 将实例标记为创建失败。
func MarkInstanceFailed(ctx context.Context, db *DB, instanceID string) (Result, error) {
	return db.Exec(ctx, "UPDATE instances SET status='failed',updated_at=now() WHERE id=$1", instanceID)
}

// FailTask 记录任务失败原因与结束时间。
func FailTask(ctx context.Context, db *DB, taskID, message string) (Result, error) {
	return db.Exec(ctx, "UPDATE instance_tasks SET status='failed',completed_at=now(),error=$2 WHERE id=$1", taskID, message)
}
