package config

import amqp "github.com/rabbitmq/amqp091-go"

const (
	RabbitMQExchange   = "fleet.events"
	RabbitMQQueue      = "geofence_alerts"
	RabbitMQDLX        = "geofence_alerts.dlx"
	RabbitMQDeadLetter = "geofence_alerts.dlq"
)

func NewRabbitMQ(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err == nil {
		err = declareTopology(ch)
	}
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(RabbitMQExchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(RabbitMQDLX, "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(RabbitMQDeadLetter, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(RabbitMQDeadLetter, "", RabbitMQDLX, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(RabbitMQQueue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": RabbitMQDLX}); err != nil {
		return err
	}
	return ch.QueueBind(RabbitMQQueue, "geofence.*", RabbitMQExchange, false, nil)
}
