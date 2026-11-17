package entity

import "time"

type VehicleLocation struct {
	ID        string    `db:"id"`
	VehicleID string    `db:"vehicle_id"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	Timestamp int64     `db:"timestamp"`
	CreatedAt time.Time `db:"created_at"`
}
