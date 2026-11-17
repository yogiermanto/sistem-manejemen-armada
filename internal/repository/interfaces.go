package repository

import (
	"context"

	"sistem-manajemen-armada/internal/entity"
)

type VehicleLocationRepository interface {
	Create(ctx context.Context, location *entity.VehicleLocation) (bool, error)
	FindLatest(ctx context.Context, vehicleID string) (*entity.VehicleLocation, error)
	FindHistory(ctx context.Context, vehicleID string, start, end int64, limit, offset int) ([]entity.VehicleLocation, error)
}

type GeofenceRepository interface {
	Create(ctx context.Context, geofence *entity.Geofence) error
	FindAll(ctx context.Context) ([]entity.Geofence, error)
}

type GeofenceEventRepository interface {
	Create(ctx context.Context, event *entity.GeofenceEvent) error
}
