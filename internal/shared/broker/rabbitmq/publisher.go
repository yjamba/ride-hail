package rabbitmq

import "context"

func (r *RMQ) Publish(ctx context.Context, queueName, queueKey string, body []byte) error {
	return nil
}
