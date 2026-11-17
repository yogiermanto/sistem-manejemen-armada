package entity

import "time"

type GeofenceEvent struct {
	ID         string    `db:"id"`
	VehicleID  string    `db:"vehicle_id"`
	GeofenceID string    `db:"geofence_id"`
	Event      string    `db:"event"`
	Latitude   float64   `db:"latitude"`
	Longitude  float64   `db:"longitude"`
	Timestamp  int64     `db:"timestamp"`
	ReceivedAt time.Time `db:"received_at"`
}
