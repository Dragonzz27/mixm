package repository

import (
	"context"
	"database/sql/driver"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WithTaskLock 在独占的物理连接上持有任务会话锁，防止重复执行。
func (db *DB) WithTaskLock(ctx context.Context, id string, fn func(*DB) error) error {
	pool, err := db.SQL()
	if err != nil {
		return err
	}

	// 固定使用一条物理连接，确保会话锁与任务操作属于同一会话。
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

	// 清理阶段单独设置超时；解锁失败时丢弃连接，避免锁泄漏。
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var unlocked bool
		if err := conn.QueryRowContext(unlockCtx, "SELECT pg_advisory_unlock(hashtextextended($1,0))", id).Scan(&unlocked); err != nil || !unlocked {
			// 标记连接失效并从连接池移除，避免仍持有锁的连接被再次使用。
			_ = conn.Raw(func(any) error {
				return driver.ErrBadConn
			})
		}
	}()

	// 为该连接构造仓储作用域，供任务内查询和事务复用。
	orm, err := gorm.Open(
		postgres.New(postgres.Config{Conn: conn}),
		&gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		return err
	}

	return fn(New(orm))
}
