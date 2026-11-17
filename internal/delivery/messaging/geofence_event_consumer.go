package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/usecase"
)

type GeofenceEventConsumer struct {
	Log         *slog.Logger
	UseCase     *usecase.GeofenceEventUseCase
	MaxAttempts int
}

func NewGeofenceEventConsumer(log *slog.Logger, useCase *usecase.GeofenceEventUseCase) *GeofenceEventConsumer {
	return &GeofenceEventConsumer{Log: log, UseCase: useCase, MaxAttempts: 3}
}

func (c *GeofenceEventConsumer) Consume(ctx context.Context, delivery amqp.Delivery) {
	event := new(model.GeofenceEvent)
	if err := json.Unmarshal(delivery.Body, event); err != nil {
		c.deadLetter(delivery, err)
		return
	}

	var err error
	for attempt := 1; attempt <= c.MaxAttempts; attempt++ {
		if err = c.UseCase.Process(ctx, event); err == nil || errors.Is(err, model.ErrBadRequest) {
			break
		}
		c.Log.Warn("process geofence event failed", "attempt", attempt, "err", err)
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}

	switch {
	case err == nil:
		delivery.Ack(false)
	case ctx.Err() != nil:
		return
	default:
		c.deadLetter(delivery, err)
	}
}

func (c *GeofenceEventConsumer) deadLetter(delivery amqp.Delivery, err error) {
	c.Log.Error("geofence event sent to DLQ", "body", string(delivery.Body), "err", err)
	delivery.Nack(false, false)
}
