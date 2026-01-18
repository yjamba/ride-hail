package ports

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type Publish interface {
	Publish(ctx context.Context, exchangeName, routingKey string, body []byte) error
}

type Consume interface {
	Consume(ctx context.Context, queueName string) (<-chan amqp091.Delivery, error)
}
