package rabbitmq

import (
	"context"

	"ride-hail/internal/shared/broker"

	"github.com/rabbitmq/amqp091-go"
)

type RMQ struct {
	Conn *amqp091.Connection
	Ch   *amqp091.Channel

	config *broker.BrokerConfig
}

func NewRMQ(config *broker.BrokerConfig) *RMQ {
	return &RMQ{
		config: config,
	}
}

func (r *RMQ) DeclareExchanges(name, kind string, durable, autoDelete, internal, noWait bool, args amqp091.Table) error {
	return r.Ch.ExchangeDeclare(
		name,
		kind,
		durable,
		autoDelete,
		internal,
		noWait,
		args,
	)
}

func (r *RMQ) DeclareQueues(queues []broker.QueueConfig) error {
	for _, queue := range queues {
		_, err := r.Ch.QueueDeclare(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			queue.Args,
		)
		if err != nil {
			return err
		}

		// Bind queue to exchange if exchange is specified
		if queue.Exchange != "" {
			err = r.Ch.QueueBind(
				queue.Name,
				queue.RoutingKey,
				queue.Exchange,
				queue.BindNoWait,
				queue.BindArgs,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *RMQ) Connect(ctx context.Context) error {
	// Create a channel to signal connection completion
	done := make(chan error, 1)

	go func() {
		conn, err := amqp091.Dial(r.config.GetConnectionString())
		if err != nil {
			done <- err
			return
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			done <- err
			return
		}

		r.Conn = conn
		r.Ch = ch
		done <- nil
	}()

	// Wait for either connection to complete or context to be cancelled
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *RMQ) Close(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		var err error

		if r.Ch != nil {
			err = r.Ch.Close()
			if err != nil {
				done <- err
				return
			}
		}

		if r.Conn != nil {
			err = r.Conn.Close()
		}

		done <- err
	}()

	// Wait for either close to complete or context to be cancelled
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *RMQ) Reconnect(ctx context.Context) error {
	// Close existing connection
	if err := r.Close(ctx); err != nil {
		return err
	}

	// Reconnect
	return r.Connect(ctx)
}
