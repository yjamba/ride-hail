package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func (r *RMQ) Publish(ctx context.Context, exchangeName, routingKey string, body []byte) error {
	if r == nil || r.Ch == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	return r.Ch.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
		},
	)
}
