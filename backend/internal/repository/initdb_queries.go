package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"mixm/backend/internal/model"
)

// SeedRole 初始化角色及其权限，已存在时保持幂等。
func SeedRole(ctx context.Context, db *DB, id, name, keys string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO roles(id,name,route_keys) VALUES($1,$2,$3) ON CONFLICT(name) DO NOTHING", id, name, keys)
}

// SeedContentType 按内容类型代码初始化默认内容类型。
func SeedContentType(ctx context.Context, db *DB, id, code, name string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO content_types(id,code,name) VALUES($1,$2,$3) ON CONFLICT(code) DO NOTHING", id, code, name)
}

// SeedExample 使用固定标识补充示例记录，已存在时保留管理员修改过的字段。
func SeedExample(ctx context.Context, db *DB, table, key string, data model.Record) (string, bool, error) {
	if Columns[table] == nil {
		return "", false, fmt.Errorf("unknown seed table: %s", table)
	}

	id := uuid.NewSHA1(uuid.NameSpaceURL, []byte("mixm:initdb:"+table+":"+key)).String()
	fields := model.Record{"id": id}
	for name, value := range data {
		if _, ok := Columns[table][name]; !ok || name == "id" {
			return "", false, fmt.Errorf("invalid seed field: %s.%s", table, name)
		}
		fields[name] = value
	}

	result := db.orm.WithContext(ctx).Table(table).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(map[string]any(fields))
	return id, result.RowsAffected > 0, result.Error
}
