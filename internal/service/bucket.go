package service

import (
	"fmt"
	"math"
	"time"

	"traffic-backend/internal/model"
)

// CreateLocationBucket converts GPS coordinates into a
// normalized location bucket for the prototype.
func CreateLocationBucket(location model.Location) string {
	latitude := math.Round(location.Latitude*10000) / 10000
	longitude := math.Round(location.Longitude*10000) / 10000

	return fmt.Sprintf("%.4f_%.4f", latitude, longitude)
}

// CreateTimeBucket places a timestamp into a 5-minute bucket.
//
// Example:
// 16:12:15 -> 16:10
// 16:14:59 -> 16:10
// 16:15:00 -> 16:15
func CreateTimeBucket(timestamp time.Time) string {
	minute := timestamp.Minute()

	bucketMinute := (minute / 5) * 5

	return fmt.Sprintf(
		"%02d:%02d",
		timestamp.Hour(),
		bucketMinute,
	)
}