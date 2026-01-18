package ports

import (
	"context"

	"ride-hail/internal/shared/broker/rabbitmq"
)

type Consume interface {
	Consume(ctx context.Context, queueName, queueKey string) (<-chan rabbitmq.Message, error)
}

type Publish interface {
	Publish(ctx context.Context, exchange, queueKey string, body []byte) error
}
