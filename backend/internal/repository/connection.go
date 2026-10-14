package repository

import (
	"context"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(ctx context.Context, dsn string) (*DB, error) {
	orm, err := gorm.Open(postgres.Open(dsn), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	db := New(orm)
	pool, err := db.SQL()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(20)
	pool.SetMaxIdleConns(20)
	pool.SetConnMaxLifetime(time.Hour)
	if err = pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return db, nil
}
