package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
)

type vehicleLocationRepository struct {
	DB *pgxpool.Pool
}

func NewVehicleLocationRepository(db *pgxpool.Pool) VehicleLocationRepository {
	return &vehicleLocationRepository{DB: db}
}

func (r *vehicleLocationRepository) Create(ctx context.Context, l *entity.VehicleLocation) (bool, error) {
	err := r.DB.QueryRow(ctx, `
		INSERT INTO vehicle_locations (id, vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (vehicle_id, timestamp) DO NOTHING
		RETURNING created_at`,
		l.ID, l.VehicleID, l.Latitude, l.Longitude, l.Timestamp).
		Scan(&l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *vehicleLocationRepository) FindLatest(ctx context.Context, vehicleID string) (*entity.VehicleLocation, error) {
	rows, _ := r.DB.Query(ctx, `
		SELECT id, vehicle_id, latitude, longitude, timestamp, created_at
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`, vehicleID)
	l, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[entity.VehicleLocation])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *vehicleLocationRepository) FindHistory(ctx context.Context, vehicleID string, start, end int64, limit, offset int) ([]entity.VehicleLocation, error) {
	rows, _ := r.DB.Query(ctx, `
		SELECT id, vehicle_id, latitude, longitude, timestamp, created_at
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
		LIMIT $4 OFFSET $5`, vehicleID, start, end, limit, offset)
	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.VehicleLocation])
}
