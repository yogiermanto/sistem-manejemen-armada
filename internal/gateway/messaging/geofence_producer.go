package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"sistem-manajemen-armada/internal/model"
)

type Dialer func() (*amqp.Connection, *amqp.Channel, error)

type GeofenceProducer struct {
	Log      *slog.Logger
	Exchange string
	dial     Dialer

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewGeofenceProducer(log *slog.Logger, exchange string, dial Dialer) (*GeofenceProducer, error) {
	p := &GeofenceProducer{Log: log, Exchange: exchange, dial: dial}
	return p, p.connect()
}

func (p *GeofenceProducer) connect() error {
	conn, ch, err := p.dial()
	if err == nil {
		err = ch.Confirm(false)
	}
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return err
	}
	p.conn, p.ch = conn, ch
	return nil
}

func routingKey(event string) string {
	return "geofence." + strings.TrimPrefix(event, "geofence_")
}

func (p *GeofenceProducer) Send(ctx context.Context, event *model.GeofenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	p.mu.Lock()
	if p.ch == nil || p.ch.IsClosed() {
		if p.conn != nil {
			p.conn.Close()
		}
		if err := p.connect(); err != nil {
			p.mu.Unlock()
			return err
		}
		p.Log.Info("rabbitmq producer reconnected")
	}
	confirm, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, p.Exchange, routingKey(event.Event), false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	p.mu.Unlock()
	if err != nil {
		return err
	}
	if ok, err := confirm.WaitContext(ctx); err != nil || !ok {
		return errors.Join(errors.New("rabbitmq did not confirm publish"), err)
	}
	return nil
}

func (p *GeofenceProducer) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		p.conn.Close()
	}
}
