package repository

import (
	"context"

	"github.com/google/uuid"
)

// SyncImageRepositories 原子补充新发现的公开仓库，名称冲突时保留已有归属、权限和关联 ID。
func SyncImageRepositories(ctx context.Context, db *DB, ownerID string, names []string) (int, error) {
	created := 0
	err := db.Transaction(ctx, func(tx *DB) error {
		for _, name := range names {
			result, err := tx.Exec(ctx, `INSERT INTO image_repositories (id, name, user_id, is_public)
				VALUES ($1, $2, $3, true) ON CONFLICT (name) DO NOTHING`, uuid.NewString(), name, ownerID)
			if err != nil {
				return err
			}

			created += int(result.RowsAffected())
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return created, nil
}
