package model

import "time"

type CreateGeofenceRequest struct {
	Name      string   `json:"name" validate:"required,max=100"`
	Latitude  *float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude *float64 `json:"longitude" validate:"required,min=-180,max=180"`
	RadiusM   float64  `json:"radius_m" validate:"omitempty,gt=0,max=100000"`
}

type GeofenceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	RadiusM   float64   `json:"radius_m"`
	CreatedAt time.Time `json:"created_at"`
}
