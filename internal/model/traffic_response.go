package model

import "time"

type FrontendTrafficResponse struct {
	EventID   string           `json:"event_id"`
	Timestamp time.Time        `json:"timestamp"`
	Location  FrontendLocation `json:"location"`
	Traffic   FrontendTraffic  `json:"traffic"`
}

type FrontendLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Area      string  `json:"area"`
	City      string  `json:"city"`
}

type FrontendTraffic struct {
	VehicleCount              int     `json:"vehicle_count"`
	HistoricalVehicleAverage  float64 `json:"historical_vehicle_average"`
	VehicleIncreasePercentage float64 `json:"vehicle_increase_percentage"`
	VehicleOccupancy          float64 `json:"vehicle_occupancy"`
	CongestionLevel           string  `json:"congestion_level"`
}
