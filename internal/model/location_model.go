package model

type LocationMessage struct {
	VehicleID string  `json:"vehicle_id" validate:"required,max=32"`
	Latitude  float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"min=-180,max=180"`
	Timestamp int64   `json:"timestamp" validate:"gt=0"`
}

type VehicleLocationResponse struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type GetLocationHistoryRequest struct {
	VehicleID string `json:"-" validate:"required,max=32"`
	Start     int64  `form:"start" validate:"required,gt=0"`
	End       int64  `form:"end" validate:"required,gtefield=Start"`
	Page      int    `form:"page,default=1" validate:"min=1"`
	Limit     int    `form:"limit,default=100" validate:"min=1,max=1000"`
}
