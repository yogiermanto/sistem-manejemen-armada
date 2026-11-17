package entity

import (
	"math"
	"testing"
)

func metersNorth(lat, m float64) float64 {
	return lat + m/(earthRadiusM*math.Pi/180)
}

func TestDistance(t *testing.T) {
	if d := Distance(-6.2088, 106.8456, metersNorth(-6.2088, 50), 106.8456); math.Abs(d-50) > 0.01 {
		t.Fatalf("want 50m, got %.4f", d)
	}

	if d := Distance(-6.1754, 106.8272, -6.1950, 106.8230); d < 2100 || d > 2300 {
		t.Fatalf("want ~2.2km, got %.0fm", d)
	}
}

func TestGeofenceContainsBoundary(t *testing.T) {
	g := Geofence{Latitude: -6.2088, Longitude: 106.8456, RadiusM: 50}
	if !g.Contains(metersNorth(g.Latitude, 49), g.Longitude) {
		t.Fatal("49m should be inside a 50m fence")
	}
	if g.Contains(metersNorth(g.Latitude, 51), g.Longitude) {
		t.Fatal("51m should be outside a 50m fence")
	}
}
