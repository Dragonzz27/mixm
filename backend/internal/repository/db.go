package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// DB scopes GORM queries to a pool, transaction, or pinned worker connection.
type DB struct{ orm *gorm.DB }

func New(orm *gorm.DB) *DB           { return &DB{orm: orm} }
func (db *DB) GORM() *gorm.DB        { return db.orm }
func (db *DB) SQL() (*sql.DB, error) { return db.orm.DB() }
func (db *DB) Close() error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}
	return pool.Close()
}
func (db *DB) Ping(ctx context.Context) error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}
	return pool.PingContext(ctx)
}
func (db *DB) Begin(ctx context.Context) (*DB, error) {
	tx := db.orm.WithContext(ctx).Begin()
	return New(tx), tx.Error
}
func (db *DB) Commit(ctx context.Context) error   { return db.orm.WithContext(ctx).Commit().Error }
func (db *DB) Rollback(ctx context.Context) error { return db.orm.WithContext(ctx).Rollback().Error }
func (db *DB) Transaction(ctx context.Context, fn func(*DB) error) error {
	return db.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(New(tx)) })
}
