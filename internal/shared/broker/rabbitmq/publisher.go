package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func (r *RMQ) Publish(ctx context.Context, exchange, key string, body []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	slog.Info("publishing message", "exchange", exchange, "key", key)

	err := r.Ch.PublishWithContext(ctx,
		exchange,
		key,
		false, // mandatory
		false, // imeddiate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	return nil
}
