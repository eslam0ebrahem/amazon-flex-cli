package utils

import (
	"math"
	"math/rand"
	"time"
)

func GenerateGPSInCircle(lat, lon, radiusMeters float64) (float64, float64) {
	rand.Seed(time.Now().UnixNano())
	radiusInDegrees := radiusMeters / 111300.0
	u := rand.Float64()
	v := rand.Float64()
	w := radiusInDegrees * math.Sqrt(u)
	t := 2 * math.Pi * v
	x := w * math.Cos(t)
	y := w * math.Sin(t)
	
	newLon := x / math.Cos(lat*math.Pi/180)
	return lat + y, lon + newLon
}
