package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
)

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
