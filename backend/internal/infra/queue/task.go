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

// process 获取任务专属锁，避免重复消息并发执行。
func process(
	ctx context.Context,
	db *repository.DB,
	cache *redis.Client,
	k *cluster.Client,
	id string,
) error {
	return db.WithTaskLock(ctx, id, func(conn *repository.DB) error {
		return processLocked(ctx, conn, cache, k, id)
	})
}

// processLocked 读取实例配置并执行部署，事务更新任务与实例状态。
func processLocked(
	ctx context.Context,
	conn *repository.DB,
	cache *redis.Client,
	k *cluster.Client,
	id string,
) error {
	// 以数据库状态为准，跳过已经完成或失败的重复消息。
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

	// 读取任务关联的实际实例记录，避免使用过期的请求快照。
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

	// 为集群部署设置独立超时，并响应 worker 退出。
	jobCtx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	defer cancel()

	provisionErr := k.Provision(jobCtx, v)
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 原子更新实例和任务状态，避免二者出现不一致。
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

	// 缓存仅加速进度查询，持久化任务仍由数据库负责。
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
