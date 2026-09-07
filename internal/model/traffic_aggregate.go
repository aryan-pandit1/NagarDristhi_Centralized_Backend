package model

import "time"

type TrafficAggregate struct {
	LocationBucket         string
	Latitude               float64
	Longitude              float64
	TimeBucket             string
	TrafficDate            time.Time
	AverageVehicleCount    float64
	AverageCongestionScore float64
	AverageOccupancy       float64
	UpdatedAt              time.Time
}
