package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
)

func channel(c config.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.DialConfig(c.RabbitURL, amqp.Config{Heartbeat: 10 * time.Second, Dial: amqp.DefaultDial(5 * time.Second)})
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	if _, err = ch.QueueDeclare(c.Queue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}
func retry(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(3 * time.Second):
		return true
	}
}
func Dispatch(ctx context.Context, c config.Config, db *repository.DB) {
	for ctx.Err() == nil {
		conn, ch, err := channel(c)
		if err != nil {
			slog.Warn("RabbitMQ publisher reconnect", "error", err)
			if !retry(ctx) {
				return
			}
			continue
		}
		err = ch.Confirm(false)
		if err != nil {
			ch.Close()
			conn.Close()
			if !retry(ctx) {
				return
			}
			continue
		}
		confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
		for ctx.Err() == nil {
			err = publishBatch(ctx, c, db, ch, confirms)
			if err != nil {
				slog.Warn("outbox publish failed", "error", err)
				break
			}
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
		}
		ch.Close()
		conn.Close()
		if !retry(ctx) {
			return
		}
	}
}
func publishBatch(ctx context.Context, c config.Config, db *repository.DB, ch *amqp.Channel, confirm <-chan amqp.Confirmation) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	rows, err := repository.OutboxTasks(ctx, tx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		id := model.String(row, "id")
		err = ch.PublishWithContext(ctx, "", c.Queue, false, false, amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: "text/plain", MessageId: id, Body: []byte(id)})
		if err != nil {
			return err
		}
		select {
		case ack, ok := <-confirm:
			if !ok || !ack.Ack {
				return fmt.Errorf("publish not confirmed")
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return fmt.Errorf("publish confirm timeout")
		}
		if _, err = repository.MarkTaskPublished(ctx, tx, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
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
