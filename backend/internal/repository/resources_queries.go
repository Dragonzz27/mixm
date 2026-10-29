package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// CategoryTreeRows 按指定顺序读取分类树的原始记录。
func CategoryTreeRows(ctx context.Context, db *DB, order, table string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM "+table+" ORDER BY "+order)
}

// CategoryContentPreview 读取分类下已发布内容的预览列表。
func CategoryContentPreview(ctx context.Context, db *DB, categoryID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT c.* FROM contents c JOIN content_categories cc ON cc.content_id=c.id WHERE cc.category_id=$1 AND c.is_active=true AND c.status='published' ORDER BY cc.sort_order LIMIT 4",
		categoryID,
	)
}

// BoundCategoryContents 读取分类绑定的全部内容。
func BoundCategoryContents(ctx context.Context, db *DB, categoryID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT c.* FROM contents c JOIN content_categories cc ON cc.content_id=c.id WHERE cc.category_id=$1 ORDER BY cc.sort_order",
		categoryID,
	)
}

// ReviewQuestion 记录题目审核结果。
func ReviewQuestion(ctx context.Context, db *DB, status, userID, questionID string) (model.Record, error) {
	return One(
		ctx,
		db,
		"UPDATE questions SET review_status=$1,reviewed_by=$2,reviewed_at=now() WHERE id=$3 RETURNING *",
		status,
		userID,
		questionID,
	)
}
