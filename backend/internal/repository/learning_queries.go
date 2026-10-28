package repository

import (
	"context"

	"mixm/backend/internal/model"
)

func LockSimulationUser(ctx context.Context, db *DB, userID string) (Result, error) {
	return db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", userID)
}
func ActiveSimulationAttempts(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM simulation_attempt WHERE user_id=$1 AND status='in_progress' ORDER BY started_at DESC LIMIT 1", userID)
}
func ContentKnowledgeTags(ctx context.Context, db *DB, contentID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT c.id,c.name FROM categories c JOIN content_categories cc ON cc.category_id=c.id WHERE cc.content_id=$1", contentID)
}
func LockSimulationUserForSubmission(ctx context.Context, db *DB, userID string) (Result, error) {
	return db.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,1))", userID)
}
func SimulationAttemptForUpdate(ctx context.Context, db *DB, attemptID, userID string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM simulation_attempt WHERE id=$1 AND user_id=$2 FOR UPDATE", attemptID, userID)
}
func CurrentSimulationAttempts(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM simulation_attempt WHERE user_id=$1 AND status='in_progress' ORDER BY started_at DESC LIMIT 1", userID)
}
func SimulationAttemptForUser(ctx context.Context, db *DB, attemptID, userID string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM simulation_attempt WHERE id=$1 AND user_id=$2", attemptID, userID)
}
