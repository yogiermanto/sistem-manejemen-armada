package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/model/converter"
	"sistem-manajemen-armada/internal/repository"
)

type GeofenceEventProducer interface {
	Send(ctx context.Context, event *model.GeofenceEvent) error
}

type VehicleLocationUseCase struct {
	Log                       *slog.Logger
	Validate                  *validator.Validate
	VehicleLocationRepository repository.VehicleLocationRepository
	Detector                  *GeofenceDetector
	GeofenceEventProducer     GeofenceEventProducer
}

func NewVehicleLocationUseCase(log *slog.Logger, validate *validator.Validate,
	vehicleLocationRepository repository.VehicleLocationRepository,
	detector *GeofenceDetector, producer GeofenceEventProducer) *VehicleLocationUseCase {
	return &VehicleLocationUseCase{
		Log:                       log,
		Validate:                  validate,
		VehicleLocationRepository: vehicleLocationRepository,
		Detector:                  detector,
		GeofenceEventProducer:     producer,
	}
}

func (c *VehicleLocationUseCase) Ingest(ctx context.Context, msg *model.LocationMessage) error {
	if err := c.Validate.Struct(msg); err != nil {
		return fmt.Errorf("%w: %v", model.ErrBadRequest, err)
	}

	location := converter.LocationMessageToEntity(msg)
	location.ID = newID()
	inserted, err := c.VehicleLocationRepository.Create(ctx, location)
	if err != nil || !inserted {
		return err
	}

	for _, event := range c.Detector.Check(location) {

		if err := c.GeofenceEventProducer.Send(ctx, &event); err != nil {
			c.Log.Error("publish geofence event", "vehicle_id", event.VehicleID, "geofence_id", event.GeofenceID, "event", event.Event, "err", err)
			continue
		}
		c.Log.Info("geofence event published", "vehicle_id", event.VehicleID, "geofence_id", event.GeofenceID, "event", event.Event)
	}
	return nil
}

func (c *VehicleLocationUseCase) GetLatest(ctx context.Context, vehicleID string) (*model.VehicleLocationResponse, error) {
	location, err := c.VehicleLocationRepository.FindLatest(ctx, vehicleID)
	if errors.Is(err, model.ErrNotFound) {
		return nil, fmt.Errorf("%w: no location for vehicle %s", model.ErrNotFound, vehicleID)
	}
	if err != nil {
		return nil, err
	}
	return converter.VehicleLocationToResponse(location), nil
}

func (c *VehicleLocationUseCase) GetHistory(ctx context.Context, req *model.GetLocationHistoryRequest) ([]model.VehicleLocationResponse, *model.PageMetadata, error) {
	if err := c.Validate.Struct(req); err != nil {
		return nil, nil, fmt.Errorf("%w: start and end are required unix seconds and end must not be before start; page min 1; limit 1-1000", model.ErrBadRequest)
	}

	locations, err := c.VehicleLocationRepository.FindHistory(ctx, req.VehicleID, req.Start, req.End, req.Limit+1, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, nil, err
	}
	hasMore := len(locations) > req.Limit
	if hasMore {
		locations = locations[:req.Limit]
	}

	responses := make([]model.VehicleLocationResponse, len(locations))
	for i := range locations {
		responses[i] = *converter.VehicleLocationToResponse(&locations[i])
	}
	return responses, &model.PageMetadata{Page: req.Page, Limit: req.Limit, Count: len(responses), HasMore: hasMore}, nil
}
