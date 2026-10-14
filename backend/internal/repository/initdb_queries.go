package repository

import (
	"context"
)

func SeedRole(ctx context.Context, db *DB, id, name, keys string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO roles(id,name,route_keys) VALUES($1,$2,$3) ON CONFLICT(name) DO NOTHING", id, name, keys)
}
func SeedContentType(ctx context.Context, db *DB, id, code, name string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO content_types(id,code,name) VALUES($1,$2,$3) ON CONFLICT(code) DO NOTHING", id, code, name)
}
