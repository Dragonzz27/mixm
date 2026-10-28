package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"mixm/backend/internal/config"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/repository"
)

func Work(ctx context.Context, c config.Config, db *repository.DB, cache *redis.Client, k *cluster.Client) error {
	if k == nil {
		return fmt.Errorf("worker requires Kubernetes configuration")
	}
	for ctx.Err() == nil {
		conn, ch, err := channel(c)
		if err != nil {
			slog.Warn("RabbitMQ consumer reconnect", "error", err)
			if !retry(ctx) {
				break
			}
			continue
		}
		if err = ch.Qos(1, 0, false); err != nil {
			ch.Close()
			conn.Close()
			if !retry(ctx) {
				break
			}
			continue
		}
		deliveries, err := ch.Consume(c.Queue, "", false, false, false, false, nil)
		if err != nil {
			ch.Close()
			conn.Close()
			if !retry(ctx) {
				break
			}
			continue
		}
		func() {
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-deliveries:
					if !ok {
						return
					}
					if err = process(ctx, db, cache, k, string(d.Body)); err != nil {
						slog.Error("task failed to persist", "error", err)
						_ = d.Nack(false, true)
						return
					}
					if err = d.Ack(false); err != nil {
						return
					}
				}
			}
		}()
		ch.Close()
		conn.Close()
		if !retry(ctx) {
			break
		}
	}
	return ctx.Err()
}
