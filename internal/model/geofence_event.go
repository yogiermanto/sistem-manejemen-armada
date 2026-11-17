package model

const (
	EventGeofenceEntry = "geofence_entry"
	EventGeofenceExit  = "geofence_exit"
)

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GeofenceEvent struct {
	VehicleID  string `json:"vehicle_id" validate:"required,max=32"`
	GeofenceID string `json:"geofence_id" validate:"required,uuid"`
	Event      string `json:"event" validate:"oneof=geofence_entry geofence_exit"`
	Location   Point  `json:"location"`
	Timestamp  int64  `json:"timestamp" validate:"gt=0"`
}
