package broker

import "github.com/rabbitmq/amqp091-go"

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp091.Table

	Exchange   string
	RoutingKey string
	BindNoWait bool
	BindArgs   amqp091.Table
}
