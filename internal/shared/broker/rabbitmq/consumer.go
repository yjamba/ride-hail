package rabbitmq

import (
	"context"
)

func (r *RMQ) Consume(ctx context.Context, exchange, key string) (<-chan Message, error) {
	deliveries, err := r.Ch.Consume(exchange, key, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	messages := make(chan Message)
	go func() {
		defer close(messages)
		for delivery := range deliveries {
			messages <- &DeliveryAdapter{delivery}
		}
	}()

	return messages, nil
}
