package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"sistem-manajemen-armada/internal/entity"
)

type geofenceEventRepository struct {
	DB *pgxpool.Pool
}

func NewGeofenceEventRepository(db *pgxpool.Pool) GeofenceEventRepository {
	return &geofenceEventRepository{DB: db}
}

func (r *geofenceEventRepository) Create(ctx context.Context, e *entity.GeofenceEvent) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO geofence_events (id, vehicle_id, geofence_id, event, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (vehicle_id, geofence_id, event, timestamp) DO NOTHING`,
		e.ID, e.VehicleID, e.GeofenceID, e.Event, e.Latitude, e.Longitude, e.Timestamp)
	return err
}
