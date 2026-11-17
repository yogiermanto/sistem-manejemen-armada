package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/usecase"
)

type LocationConsumer struct {
	ctx     context.Context
	Log     *slog.Logger
	UseCase *usecase.VehicleLocationUseCase
}

func NewLocationConsumer(ctx context.Context, log *slog.Logger, useCase *usecase.VehicleLocationUseCase) *LocationConsumer {
	return &LocationConsumer{ctx: ctx, Log: log, UseCase: useCase}
}

func (c *LocationConsumer) Handle(_ mqtt.Client, raw mqtt.Message) {
	message := new(model.LocationMessage)
	if err := json.Unmarshal(raw.Payload(), message); err != nil {
		c.drop(raw, "invalid json", err)
		return
	}
	if topicVehicleID(raw.Topic()) != message.VehicleID {
		c.drop(raw, "vehicle_id does not match topic", nil)
		return
	}

	for {
		err := c.UseCase.Ingest(c.ctx, message)
		if err == nil {
			raw.Ack()
			return
		}
		if errors.Is(err, model.ErrBadRequest) {
			c.drop(raw, "invalid location", err)
			return
		}

		c.Log.Error("ingest location, retrying", "vehicle_id", message.VehicleID, "err", err)
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (c *LocationConsumer) drop(raw mqtt.Message, reason string, err error) {
	c.Log.Warn("drop location", "topic", raw.Topic(), "reason", reason, "err", err)
	raw.Ack()
}

func topicVehicleID(topic string) string {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) != 4 {
		return ""
	}
	return parts[2]
}
