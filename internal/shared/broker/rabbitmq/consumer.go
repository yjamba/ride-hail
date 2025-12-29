package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

func (r *RMQ) Consume(ctx context.Context, queueName, queueKey string) (<-chan amqp091.Delivery, error) {
	return nil, nil
}
