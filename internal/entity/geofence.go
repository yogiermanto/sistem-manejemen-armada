package entity

import (
	"math"
	"time"
)

type Geofence struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	RadiusM   float64   `db:"radius_m"`
	CreatedAt time.Time `db:"created_at"`
}

func (g Geofence) Contains(lat, lon float64) bool {
	return Distance(lat, lon, g.Latitude, g.Longitude) <= g.RadiusM
}

const earthRadiusM = 6371000.0

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	a := math.Pow(math.Sin(dLat/2), 2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Pow(math.Sin(dLon/2), 2)
	return 2 * earthRadiusM * math.Asin(math.Sqrt(a))
}
