package app

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/repository"
)

func Open(ctx context.Context) (config.Config, *repository.DB, *redis.Client, error) {
	c, err := config.Load()
	if err != nil {
		return c, nil, nil, err
	}
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
	opts, err := redis.ParseURL(c.RedisURL)
	if err != nil {
		db.Close()
		return c, nil, nil, fmt.Errorf("REDIS_URL: %w", err)
	}
	return c, db, redis.NewClient(opts), nil
}
