package usecase

import (
	"fmt"
	"math"
	"testing"

	"sistem-manajemen-armada/internal/entity"
	"sistem-manajemen-armada/internal/model"
)

var defaultFence = entity.Geofence{ID: "g1", Name: "Default", Latitude: -6.2088, Longitude: 106.8456, RadiusM: 50}

func north(vehicle string, m float64, ts int64) *entity.VehicleLocation {
	return &entity.VehicleLocation{
		VehicleID: vehicle,
		Latitude:  defaultFence.Latitude + m/(6371000.0*math.Pi/180),
		Longitude: defaultFence.Longitude,
		Timestamp: ts,
	}
}

func eventNames(evs []model.GeofenceEvent) string {
	s := ""
	for _, e := range evs {
		s += fmt.Sprintf("%s@%s ", e.Event, e.GeofenceID)
	}
	return s
}

func TestDetectorEntryExit(t *testing.T) {
	d := NewGeofenceDetector([]entity.Geofence{defaultFence})
	steps := []struct {
		meters float64
		ts     int64
		want   string
	}{
		{100, 1, ""},
		{51, 2, ""},
		{49, 3, "geofence_entry@g1 "},
		{10, 4, ""},
		{0, 3, ""},
		{80, 5, "geofence_exit@g1 "},
		{90, 6, ""},
	}
	for _, s := range steps {
		if got := eventNames(d.Check(north("B1", s.meters, s.ts))); got != s.want {
			t.Fatalf("ping %vm ts=%d: want %q, got %q", s.meters, s.ts, s.want, got)
		}
	}
}

func TestDetectorFirstPingInside(t *testing.T) {
	d := NewGeofenceDetector([]entity.Geofence{defaultFence})
	if got := eventNames(d.Check(north("B2", 0, 1))); got != "geofence_entry@g1 " {
		t.Fatalf("want entry on first ping inside, got %q", got)
	}
}

func TestDetectorMultipleAndAddedFences(t *testing.T) {
	d := NewGeofenceDetector([]entity.Geofence{defaultFence})
	d.Check(north("B3", 300, 1))

	p := north("", 200, 0)
	d.Add(entity.Geofence{ID: "g2", Latitude: p.Latitude, Longitude: p.Longitude, RadiusM: 30})

	if got := eventNames(d.Check(north("B3", 190, 2))); got != "geofence_entry@g2 " {
		t.Fatalf("want entry into added fence 2, got %q", got)
	}
	if got := eventNames(d.Check(north("B3", 0, 3))); got != "geofence_entry@g1 geofence_exit@g2 " {
		t.Fatalf("want entry 1 + exit 2 in one ping, got %q", got)
	}

	if got := eventNames(d.Check(north("B4", 0, 1))); got != "geofence_entry@g1 " {
		t.Fatalf("want independent state for B4, got %q", got)
	}
}
