package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func (r *RMQ) Consume(ctx context.Context, queueName string) (<-chan amqp091.Delivery, error) {
	if r == nil || r.Ch == nil {
		return nil, fmt.Errorf("rabbitmq channel is not initialized")
	}

	if err := r.Ch.Qos(10, 0, false); err != nil {
		return nil, err
	}

	return r.Ch.ConsumeWithContext(
		ctx,
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}
