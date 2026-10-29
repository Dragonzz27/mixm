package queue

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"mixm/backend/internal/config"
)

// channel 连接 RabbitMQ 并声明持久化实例任务队列。
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

// retry 等待重试间隔，支持通过上下文取消。
func retry(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(3 * time.Second):
		return true
	}
}
