package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// LockSimulationUser 获取用户测评事务锁，避免并发创建测试。
func LockSimulationUser(ctx context.Context, db *DB, userID string) (Result, error) {
	return db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", userID)
}

// ActiveSimulationAttempts 读取用户最新的进行中测评。
func ActiveSimulationAttempts(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT * FROM simulation_attempt WHERE user_id=$1 AND status='in_progress' ORDER BY started_at DESC LIMIT 1",
		userID,
	)
}

// ContentKnowledgeTags 读取内容关联的知识分类标签。
func ContentKnowledgeTags(ctx context.Context, db *DB, contentID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT c.id,c.name FROM categories c JOIN content_categories cc ON cc.category_id=c.id WHERE cc.content_id=$1",
		contentID,
	)
}

// LockSimulationUserForSubmission 获取用户测评事务锁，串行处理答卷提交。
func LockSimulationUserForSubmission(ctx context.Context, db *DB, userID string) (Result, error) {
	return db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", userID)
}

// SimulationAttemptForUpdate 锁定指定用户的测评记录以便更新。
func SimulationAttemptForUpdate(ctx context.Context, db *DB, attemptID, userID string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM simulation_attempt WHERE id=$1 AND user_id=$2 FOR UPDATE", attemptID, userID)
}

// CurrentSimulationAttempts 查询用户当前进行中的测评记录。
func CurrentSimulationAttempts(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT * FROM simulation_attempt WHERE user_id=$1 AND status='in_progress' ORDER BY started_at DESC LIMIT 1",
		userID,
	)
}

// SimulationAttemptForUser 读取属于指定用户的测评记录。
func SimulationAttemptForUser(ctx context.Context, db *DB, attemptID, userID string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM simulation_attempt WHERE id=$1 AND user_id=$2", attemptID, userID)
}
