package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
)

// Message — абстрактный интерфейс для сообщения из брокера
type Message interface {
	Body() []byte
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
}

// DeliveryAdapter адаптирует amqp091.Delivery к интерфейсу Message
type DeliveryAdapter struct {
	delivery amqp091.Delivery
}

func (d *DeliveryAdapter) Body() []byte {
	return d.delivery.Body
}

func (d *DeliveryAdapter) Ack(multiple bool) error {
	return d.delivery.Ack(multiple)
}

func (d *DeliveryAdapter) Nack(multiple, requeue bool) error {
	return d.delivery.Nack(multiple, requeue)
}
