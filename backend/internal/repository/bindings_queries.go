package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// ClearContentBindings 删除指定内容的全部分类绑定。
func ClearContentBindings(ctx context.Context, db *DB, contentID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM content_categories WHERE content_id=$1", contentID)
}

// BindCategoryContent 建立分类与内容的绑定关系。
func BindCategoryContent(ctx context.Context, db *DB, id, contentID, categoryID string) (Result, error) {
	return db.Exec(
		ctx,
		"INSERT INTO content_categories(id,content_id,category_id) VALUES($1,$2,$3) ON CONFLICT(content_id,category_id) DO NOTHING",
		id,
		contentID,
		categoryID,
	)
}

// UnbindCategoryContent 移除分类与内容的指定绑定关系。
func UnbindCategoryContent(ctx context.Context, db *DB, categoryID, contentID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM content_categories WHERE category_id=$1 AND content_id=$2", categoryID, contentID)
}

// BindingCounts 统计指定分类下的内容绑定数量。
func BindingCounts(ctx context.Context, db *DB, categoryIDs []string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT category_id,count(*) AS n FROM content_categories WHERE category_id=ANY($1) GROUP BY category_id",
		categoryIDs,
	)
}

// ActiveCategories 读取启用的分类。
func ActiveCategories(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(ctx, db, "SELECT * FROM categories WHERE is_active=true ORDER BY sort_order,id")
}
