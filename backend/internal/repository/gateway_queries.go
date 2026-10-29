package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// RunningSSHInstance 根据 SSH 用户名查找运行中的实例。
func RunningSSHInstance(ctx context.Context, db *DB, username string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM instances WHERE ssh_username=$1 AND status='running'", username)
}
