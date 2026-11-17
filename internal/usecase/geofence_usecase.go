package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"

	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/model/converter"
	"sistem-manajemen-armada/internal/repository"
)

const defaultGeofenceRadiusM = 50

type GeofenceUseCase struct {
	Log                *slog.Logger
	Validate           *validator.Validate
	GeofenceRepository repository.GeofenceRepository
	Detector           *GeofenceDetector
}

func NewGeofenceUseCase(log *slog.Logger, validate *validator.Validate,
	geofenceRepository repository.GeofenceRepository, detector *GeofenceDetector) *GeofenceUseCase {
	return &GeofenceUseCase{
		Log:                log,
		Validate:           validate,
		GeofenceRepository: geofenceRepository,
		Detector:           detector,
	}
}

func (c *GeofenceUseCase) Create(ctx context.Context, req *model.CreateGeofenceRequest) (*model.GeofenceResponse, error) {
	if err := c.Validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: name (max 100 chars), latitude (-90..90) and longitude (-180..180) are required; radius_m optional, 0-100000, default 50", model.ErrBadRequest)
	}
	if req.RadiusM == 0 {
		req.RadiusM = defaultGeofenceRadiusM
	}

	geofence := &entity.Geofence{ID: newID(), Name: req.Name, Latitude: *req.Latitude, Longitude: *req.Longitude, RadiusM: req.RadiusM}
	if err := c.GeofenceRepository.Create(ctx, geofence); err != nil {
		return nil, err
	}
	c.Detector.Add(*geofence)
	c.Log.Info("geofence created", "geofence_id", geofence.ID, "name", geofence.Name, "radius_m", geofence.RadiusM)
	return converter.GeofenceToResponse(geofence), nil
}

func (c *GeofenceUseCase) List(ctx context.Context) ([]model.GeofenceResponse, error) {
	geofences, err := c.GeofenceRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]model.GeofenceResponse, len(geofences))
	for i := range geofences {
		responses[i] = *converter.GeofenceToResponse(&geofences[i])
	}
	return responses, nil
}
