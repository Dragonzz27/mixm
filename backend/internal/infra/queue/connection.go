package queue

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"mixm/backend/internal/config"
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
