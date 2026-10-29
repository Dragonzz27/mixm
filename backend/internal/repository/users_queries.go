package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// AllRoles 读取全部角色及其路由权限。
func AllRoles(ctx context.Context, db *DB) ([]model.Record, error) {
	return Query(ctx, db, "SELECT id,name FROM roles")
}

// LoginUser 按登录标识查询用户凭据。
func LoginUser(ctx context.Context, db *DB, email string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM users WHERE email=$1 AND is_active=true", email)
}

// UserRoles 查询用户关联的角色信息。
func UserRoles(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(ctx, db, "SELECT r.* FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1", userID)
}

// RefreshUser 读取刷新令牌对应的有效用户。
func RefreshUser(ctx context.Context, db *DB, username string) (model.Record, error) {
	return One(ctx, db, "SELECT * FROM users WHERE username=$1 AND is_active=true", username)
}

// UserRoleSummary 汇总用户角色及权限信息。
func UserRoleSummary(ctx context.Context, db *DB, userID string) ([]model.Record, error) {
	return Query(
		ctx,
		db,
		"SELECT r.id,r.name,r.route_keys FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1",
		userID,
	)
}

// ClearUserRoles 删除用户已有的角色关联。
func ClearUserRoles(ctx context.Context, db *DB, userID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM user_roles WHERE user_id=$1", userID)
}
