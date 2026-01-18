package ports

import "ride-hail/internal/shared/broker/rabbitmq"

type Consume interface {
	Consume(queueName, queueKey string) (<-chan rabbitmq.Message, error)
}

type Publish interface {
	Publish(exchange, queueKey string, body []byte) error
}
