package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sistem-manajemen-armada/internal/entity"
)

type geofenceRepository struct {
	DB *pgxpool.Pool
}

func NewGeofenceRepository(db *pgxpool.Pool) GeofenceRepository {
	return &geofenceRepository{DB: db}
}

func (r *geofenceRepository) Create(ctx context.Context, g *entity.Geofence) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO geofences (id, name, latitude, longitude, radius_m)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		g.ID, g.Name, g.Latitude, g.Longitude, g.RadiusM).
		Scan(&g.CreatedAt)
}

func (r *geofenceRepository) FindAll(ctx context.Context) ([]entity.Geofence, error) {
	rows, _ := r.DB.Query(ctx, `
		SELECT id, name, latitude, longitude, radius_m, created_at
		FROM geofences
		ORDER BY created_at, id`)
	return pgx.CollectRows(rows, pgx.RowToStructByName[entity.Geofence])
}
