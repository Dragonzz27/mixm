package repository

import (
	"context"

	"mixm-edu/backend/internal/model"
)

func CategoryTreeRows(ctx context.Context, db *DB, order, table string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM "+table+" ORDER BY "+order)
}
func CategoryContentPreview(ctx context.Context, db *DB, categoryID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT c.* FROM contents c JOIN content_categories cc ON cc.content_id=c.id WHERE cc.category_id=$1 AND c.is_active=true AND c.status='published' ORDER BY cc.sort_order LIMIT 4", categoryID)
}
func BoundCategoryContents(ctx context.Context, db *DB, categoryID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT c.* FROM contents c JOIN content_categories cc ON cc.content_id=c.id WHERE cc.category_id=$1 ORDER BY cc.sort_order", categoryID)
}
func ReviewQuestion(ctx context.Context, db *DB, status, userID, questionID string) (model.Record, error) {
	return One(ctx, db, "UPDATE questions SET review_status=$1,reviewed_by=$2,reviewed_at=now() WHERE id=$3 RETURNING *", status, userID, questionID)
}
