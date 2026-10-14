package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

func Open(ctx context.Context) (config.Config, *pgxpool.Pool, *redis.Client, error) {
	c, e := config.Load()
	if e != nil {
		return c, nil, nil, e
	}
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	db, e := store.Open(startup, c.DatabaseURL)
	if e != nil {
		return c, nil, nil, e
	}
	if e = store.Migrate(startup, db); e != nil {
		db.Close()
		return c, nil, nil, e
	}
	opts, e := redis.ParseURL(c.RedisURL)
	if e != nil {
		db.Close()
		return c, nil, nil, fmt.Errorf("REDIS_URL: %w", e)
	}
	return c, db, redis.NewClient(opts), nil
}
