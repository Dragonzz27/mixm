package repository

import (
	"context"
)

// SeedRole 初始化角色及其权限，已存在时保持幂等。
func SeedRole(ctx context.Context, db *DB, id, name, keys string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO roles(id,name,route_keys) VALUES($1,$2,$3) ON CONFLICT(name) DO NOTHING", id, name, keys)
}

// SeedContentType 按内容类型代码初始化默认内容类型。
func SeedContentType(ctx context.Context, db *DB, id, code, name string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO content_types(id,code,name) VALUES($1,$2,$3) ON CONFLICT(code) DO NOTHING", id, code, name)
}
