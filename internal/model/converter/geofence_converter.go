package converter

import (
	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
)

func GeofenceToResponse(g *entity.Geofence) *model.GeofenceResponse {
	return &model.GeofenceResponse{
		ID:        g.ID,
		Name:      g.Name,
		Latitude:  g.Latitude,
		Longitude: g.Longitude,
		RadiusM:   g.RadiusM,
		CreatedAt: g.CreatedAt,
	}
}

func GeofenceEventToEntity(e *model.GeofenceEvent) *entity.GeofenceEvent {
	return &entity.GeofenceEvent{
		VehicleID:  e.VehicleID,
		GeofenceID: e.GeofenceID,
		Event:      e.Event,
		Latitude:   e.Location.Latitude,
		Longitude:  e.Location.Longitude,
		Timestamp:  e.Timestamp,
	}
}
