package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// ActiveUser 根据用户名读取启用状态的用户。
func ActiveUser(ctx context.Context, db *DB, username string) (model.User, error) {
	var user model.User
	err := db.orm.WithContext(ctx).Where("username = ? AND is_active = true", username).First(&user).Error

	return user, err
}
