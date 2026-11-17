package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/model/converter"
	"sistem-manajemen-armada/internal/repository"
)

type GeofenceEventUseCase struct {
	Log                     *slog.Logger
	Validate                *validator.Validate
	GeofenceEventRepository repository.GeofenceEventRepository
}

func NewGeofenceEventUseCase(log *slog.Logger, validate *validator.Validate,
	geofenceEventRepository repository.GeofenceEventRepository) *GeofenceEventUseCase {
	return &GeofenceEventUseCase{Log: log, Validate: validate, GeofenceEventRepository: geofenceEventRepository}
}

func (c *GeofenceEventUseCase) Process(ctx context.Context, event *model.GeofenceEvent) error {
	if err := c.Validate.Struct(event); err != nil {
		return fmt.Errorf("%w: %v", model.ErrBadRequest, err)
	}
	row := converter.GeofenceEventToEntity(event)
	row.ID = newID()
	if err := c.GeofenceEventRepository.Create(ctx, row); err != nil {
		return err
	}
	c.Log.Info("geofence event processed", "vehicle_id", event.VehicleID, "geofence_id", event.GeofenceID, "event", event.Event,
		"latitude", event.Location.Latitude, "longitude", event.Location.Longitude, "timestamp", event.Timestamp)
	return nil
}
