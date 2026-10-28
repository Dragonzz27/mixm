package repository

import (
	"context"

	"mixm/backend/internal/model"
)

func AllRoles(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(ctx, db, "SELECT id,name FROM roles")
}
func LoginUser(ctx context.Context, db *DB, email string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM users WHERE email=$1 AND is_active=true", email)
}
func UserRoles(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT r.* FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1", userID)
}
func RefreshUser(ctx context.Context, db *DB, username string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM users WHERE username=$1 AND is_active=true", username)
}
func UserRoleSummary(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT r.id,r.name,r.route_keys FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1", userID)
}
func ClearUserRoles(ctx context.Context, db *DB, userID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM user_roles WHERE user_id=$1", userID)
}
