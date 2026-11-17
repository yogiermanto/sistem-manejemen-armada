package converter

import (
	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
)

func LocationMessageToEntity(m *model.LocationMessage) *entity.VehicleLocation {
	return &entity.VehicleLocation{
		VehicleID: m.VehicleID,
		Latitude:  m.Latitude,
		Longitude: m.Longitude,
		Timestamp: m.Timestamp,
	}
}

func VehicleLocationToResponse(l *entity.VehicleLocation) *model.VehicleLocationResponse {
	return &model.VehicleLocationResponse{
		VehicleID: l.VehicleID,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Timestamp: l.Timestamp,
	}
}
