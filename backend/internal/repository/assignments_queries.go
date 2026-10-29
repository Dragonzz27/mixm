package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// SubmissionExport 读取用于导出作业提交记录的数据。
func SubmissionExport(ctx context.Context, db *DB, assignmentID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT s.*,u.username FROM submissions s JOIN users u ON u.id=s.user_id WHERE s.assignment_id=$1 ORDER BY s.submitted_at",
		assignmentID,
	)
}
