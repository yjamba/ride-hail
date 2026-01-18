package ports

type Consume interface {
	Consume(queueName, queueKey string) (<-chan []byte, error)
}

type Publish interface {
	Publish(exchange, queueKey string, body []byte) error
}
