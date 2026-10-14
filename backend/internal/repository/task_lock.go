package repository

import (
	"context"
	"database/sql/driver"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WithTaskLock holds a session advisory lock on one dedicated physical connection.
func (db *DB) WithTaskLock(ctx context.Context, id string, fn func(*DB) error) error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}
	conn, err := pool.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var locked bool
	if err = conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtextextended($1,0))", id).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var unlocked bool
		if err := conn.QueryRowContext(unlockCtx, "SELECT pg_advisory_unlock(hashtextextended($1,0))", id).Scan(&unlocked); err != nil || !unlocked {
			// ErrBadConn removes the locked connection instead of returning it to the pool.
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
	}()
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	}
	return fn(New(orm))
}
