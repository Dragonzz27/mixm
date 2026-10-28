package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

func process(ctx context.Context, db *repository.DB, cache *redis.Client, k *cluster.Client, id string) error {
	return db.WithTaskLock(ctx, id, func(conn *repository.DB) error { return processLocked(ctx, conn, cache, k, id) })
}
func processLocked(ctx context.Context, conn *repository.DB, cache *redis.Client, k *cluster.Client, id string) error {
	task, err := repository.Get(ctx, conn, "instance_tasks", id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil
		}
		return err
	}
	status := model.String(task, "status")
	if status == "completed" || status == "failed" {
		return nil
	}
	instanceID := model.String(task, "instance_id")
	if instanceID == "" {
		return fmt.Errorf("task has no instance")
	}
	v, err := repository.Get(ctx, conn, "instances", instanceID)
	if err != nil {
		return err
	}
	if _, err = repository.MarkTaskProcessing(ctx, conn, id); err != nil {
		return err
	}
	jobCtx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	defer cancel()
	provisionErr := k.Provision(jobCtx, v)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if provisionErr == nil {
		_, err = repository.MarkInstanceRunning(ctx, tx, instanceID)
		if err == nil {
			_, err = repository.CompleteTask(ctx, tx, id)
		}
	} else {
		_, err = repository.MarkInstanceFailed(ctx, tx, instanceID)
		if err == nil {
			_, err = repository.FailTask(ctx, tx, id, provisionErr.Error())
		}
	}
	if err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if cache != nil {
		fresh, err := repository.Get(ctx, conn, "instance_tasks", id)
		if err == nil {
			b, _ := json.Marshal(fresh)
			if err = cache.Set(ctx, "instance_task:"+id, b, 24*time.Hour).Err(); err != nil {
				slog.Warn("task cache unavailable", "error", err)
			}
		}
	}
	return nil
}
