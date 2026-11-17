package usecase

import (
	"sync"

	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
)

type insideKey struct {
	vehicleID  string
	geofenceID string
}

type GeofenceDetector struct {
	mu     sync.Mutex
	fences []entity.Geofence
	inside map[insideKey]bool
	lastTS map[string]int64
}

func NewGeofenceDetector(fences []entity.Geofence) *GeofenceDetector {
	return &GeofenceDetector{fences: fences, inside: map[insideKey]bool{}, lastTS: map[string]int64{}}
}

func (d *GeofenceDetector) Add(f entity.Geofence) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.fences = append(d.fences, f)
}

func (d *GeofenceDetector) Check(l *entity.VehicleLocation) []model.GeofenceEvent {
	d.mu.Lock()
	defer d.mu.Unlock()
	if last, seen := d.lastTS[l.VehicleID]; seen && l.Timestamp <= last {
		return nil
	}
	d.lastTS[l.VehicleID] = l.Timestamp

	var events []model.GeofenceEvent
	for _, f := range d.fences {
		key := insideKey{l.VehicleID, f.ID}
		now := f.Contains(l.Latitude, l.Longitude)
		if now == d.inside[key] {
			continue
		}
		event := model.EventGeofenceExit
		if now {
			event = model.EventGeofenceEntry
			d.inside[key] = true
		} else {
			delete(d.inside, key)
		}
		events = append(events, model.GeofenceEvent{
			VehicleID:  l.VehicleID,
			GeofenceID: f.ID,
			Event:      event,
			Location:   model.Point{Latitude: l.Latitude, Longitude: l.Longitude},
			Timestamp:  l.Timestamp,
		})
	}
	return events
}
