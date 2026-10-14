package repository

import (
	"context"

	"mixm-edu/backend/internal/model"
)

func RunningSSHInstance(ctx context.Context, db *DB, username string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM instances WHERE ssh_username=$1 AND status='running'", username)
}
