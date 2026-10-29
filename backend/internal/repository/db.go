package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// DB 将查询限定在连接池、事务或任务专属连接内。
type DB struct{ orm *gorm.DB }

// New 将 GORM 连接包装为仓储作用域。
func New(orm *gorm.DB) *DB {
	return &DB{orm: orm}
}

// GORM 返回当前作用域内的 GORM 连接。
func (db *DB) GORM() *gorm.DB {
	return db.orm
}

// SQL 获取底层 SQL 连接池。
func (db *DB) SQL() (*sql.DB, error) {
	return db.orm.DB()
}

// Close 关闭当前数据库连接池。
func (db *DB) Close() error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}

	return pool.Close()
}

// Ping 检查数据库连接是否可用。
func (db *DB) Ping(ctx context.Context) error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}

	return pool.PingContext(ctx)
}

// Begin 创建一个数据库事务作用域。
func (db *DB) Begin(ctx context.Context) (*DB, error) {
	tx := db.orm.WithContext(ctx).Begin()
	return New(tx), tx.Error
}

// Commit 提交当前数据库事务。
func (db *DB) Commit(ctx context.Context) error {
	return db.orm.WithContext(ctx).Commit().Error
}

// Rollback 回滚当前数据库事务。
func (db *DB) Rollback(ctx context.Context) error {
	return db.orm.WithContext(ctx).Rollback().Error
}

// Transaction 在独立事务中执行回调并处理提交或回滚。
func (db *DB) Transaction(ctx context.Context, fn func(*DB) error) error {
	return db.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(New(tx))
	})
}
