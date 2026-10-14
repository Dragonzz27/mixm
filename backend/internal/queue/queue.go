package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

func channel(c config.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, e := amqp.DialConfig(c.RabbitURL, amqp.Config{Heartbeat: 10 * time.Second, Dial: amqp.DefaultDial(5 * time.Second)})
	if e != nil {
		return nil, nil, e
	}
	ch, e := conn.Channel()
	if e != nil {
		conn.Close()
		return nil, nil, e
	}
	if _, e = ch.QueueDeclare(c.Queue, true, false, false, false, nil); e != nil {
		ch.Close()
		conn.Close()
		return nil, nil, e
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
func Dispatch(ctx context.Context, c config.Config, db *pgxpool.Pool) {
	for ctx.Err() == nil {
		conn, ch, e := channel(c)
		if e != nil {
			slog.Warn("RabbitMQ publisher reconnect", "error", e)
			if !retry(ctx) {
				return
			}
			continue
		}
		e = ch.Confirm(false)
		if e != nil {
			ch.Close()
			conn.Close()
			if !retry(ctx) {
				return
			}
			continue
		}
		confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
		for ctx.Err() == nil {
			e = publishBatch(ctx, c, db, ch, confirms)
			if e != nil {
				slog.Warn("outbox publish failed", "error", e)
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
func publishBatch(ctx context.Context, c config.Config, db *pgxpool.Pool, ch *amqp.Channel, confirm <-chan amqp.Confirmation) error {
	tx, e := db.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	rows, e := store.Query(ctx, tx, "SELECT id FROM instance_tasks WHERE status IN ('pending','processing') AND (published_at IS NULL OR published_at<now()-interval '10 minutes') ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED")
	if e != nil {
		return e
	}
	for _, row := range rows {
		id := store.String(row, "id")
		e = ch.PublishWithContext(ctx, "", c.Queue, false, false, amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: "text/plain", MessageId: id, Body: []byte(id)})
		if e != nil {
			return e
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
		if _, e = tx.Exec(ctx, "UPDATE instance_tasks SET published_at=now() WHERE id=$1", id); e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
func Work(ctx context.Context, c config.Config, db *pgxpool.Pool, cache *redis.Client, k *cluster.Client) error {
	if k == nil {
		return fmt.Errorf("worker requires Kubernetes configuration")
	}
	for ctx.Err() == nil {
		conn, ch, e := channel(c)
		if e != nil {
			slog.Warn("RabbitMQ consumer reconnect", "error", e)
			if !retry(ctx) {
				break
			}
			continue
		}
		if e = ch.Qos(1, 0, false); e != nil {
			ch.Close()
			conn.Close()
			if !retry(ctx) {
				break
			}
			continue
		}
		deliveries, e := ch.Consume(c.Queue, "", false, false, false, false, nil)
		if e != nil {
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
					if e = process(ctx, db, cache, k, string(d.Body)); e != nil {
						slog.Error("task failed to persist", "error", e)
						_ = d.Nack(false, true)
						return
					}
					if e = d.Ack(false); e != nil {
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
func process(ctx context.Context, db *pgxpool.Pool, cache *redis.Client, k *cluster.Client, id string) error {
	conn, e := db.Acquire(ctx)
	if e != nil {
		return e
	}
	defer conn.Release()
	var locked bool
	if e = conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtextextended($1,0))", id).Scan(&locked); e != nil {
		return e
	}
	if !locked {
		return nil
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.Exec(unlockCtx, "SELECT pg_advisory_unlock(hashtextextended($1,0))", id); err != nil {
			_ = conn.Conn().Close(unlockCtx)
		}
	}()
	task, e := store.Get(ctx, conn, "instance_tasks", id)
	if e != nil {
		if e == pgx.ErrNoRows {
			return nil
		}
		return e
	}
	status := store.String(task, "status")
	if status == "completed" || status == "failed" {
		return nil
	}
	instanceID := store.String(task, "instance_id")
	if instanceID == "" {
		return fmt.Errorf("task has no instance")
	}
	v, e := store.Get(ctx, conn, "instances", instanceID)
	if e != nil {
		return e
	}
	if _, e = conn.Exec(ctx, "UPDATE instance_tasks SET status='processing',started_at=COALESCE(started_at,now()),attempts=attempts+1 WHERE id=$1", id); e != nil {
		return e
	}
	jobCtx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	defer cancel()
	provisionErr := k.Provision(jobCtx, v)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	tx, e := conn.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if provisionErr == nil {
		_, e = tx.Exec(ctx, "UPDATE instances SET status='running',started_at=COALESCE(started_at,now()),updated_at=now() WHERE id=$1", instanceID)
		if e == nil {
			_, e = tx.Exec(ctx, "UPDATE instance_tasks SET status='completed',completed_at=now(),error=NULL WHERE id=$1", id)
		}
	} else {
		_, e = tx.Exec(ctx, "UPDATE instances SET status='failed',updated_at=now() WHERE id=$1", instanceID)
		if e == nil {
			_, e = tx.Exec(ctx, "UPDATE instance_tasks SET status='failed',completed_at=now(),error=$2 WHERE id=$1", id, provisionErr.Error())
		}
	}
	if e != nil {
		return e
	}
	if e = tx.Commit(ctx); e != nil {
		return e
	}
	if cache != nil {
		fresh, e := store.Get(ctx, conn, "instance_tasks", id)
		if e == nil {
			b, _ := json.Marshal(fresh)
			if e = cache.Set(ctx, "instance_task:"+id, b, 24*time.Hour).Err(); e != nil {
				slog.Warn("task cache unavailable", "error", e)
			}
		}
	}
	return nil
}
