package model

import "time"

type TrafficEvent struct {
	EventID   string       `json:"event_id"`
	BusID     string       `json:"bus_id"`
	Timestamp time.Time    `json:"timestamp"`
	Location  Location     `json:"location"`
	Traffic   Traffic      `json:"traffic"`
	Vehicles  Vehicles     `json:"vehicles"`
	Source    Source       `json:"source"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Traffic struct {
	VehicleCount     int     `json:"vehicle_count"`
	VehicleOccupancy float64 `json:"vehicle_occupancy"`
}

type Vehicles struct {
	Car        int `json:"car"`
	Motorcycle int `json:"motorcycle"`
	Bus        int `json:"bus"`
	Truck      int `json:"truck"`
}

type Source struct {
	Type    string `json:"type"`
	VideoID string `json:"video_id"`
}
