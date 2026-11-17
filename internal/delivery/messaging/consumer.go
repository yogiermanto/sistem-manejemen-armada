package messaging

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DeliveryHandler func(ctx context.Context, delivery amqp.Delivery)

func ConsumeQueue(ctx context.Context, conn *amqp.Connection, deliveries <-chan amqp.Delivery, handler DeliveryHandler) error {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	for delivery := range deliveries {
		handler(ctx, delivery)
	}
	if ctx.Err() != nil {
		return nil
	}
	return errors.New("rabbitmq connection closed")
}
