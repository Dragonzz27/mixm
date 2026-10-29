package app

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"mixm/backend/internal/config"
	"mixm/backend/internal/repository"
)

// Open 读取环境配置，初始化数据库结构并创建缓存连接。
func Open(ctx context.Context) (config.Config, *repository.DB, *redis.Client, error) {
	// 集中读取环境配置，供 API 和 worker 共用。
	c, err := config.Load()
	if err != nil {
		return c, nil, nil, err
	}

	// 为数据库连接与结构初始化设置启动超时。
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	db, err := repository.Open(startup, c.DatabaseURL)
	if err != nil {
		return c, nil, nil, err
	}

	if err = repository.Migrate(startup, db); err != nil {
		db.Close()
		return c, nil, nil, err
	}

	// 解析缓存连接地址并创建 Redis 客户端。
	opts, err := redis.ParseURL(c.RedisURL)
	if err != nil {
		db.Close()
		return c, nil, nil, fmt.Errorf("REDIS_URL: %w", err)
	}

	return c, db, redis.NewClient(opts), nil
}
