package service

import (
	"context"

	"traffic-backend/internal/model"
	"traffic-backend/internal/repository"
)

const (
	MotorcycleWeight = 1
	CarWeight        = 2
	TruckWeight      = 3
	BusWeight        = 3
)

func CalculateVehicleCount(v model.Vehicles) int {
	return v.Motorcycle +
		v.Car +
		v.Truck +
		v.Bus
}

func CalculateCongestionScore(v model.Vehicles) int {
	return v.Motorcycle*MotorcycleWeight +
		v.Car*CarWeight +
		v.Truck*TruckWeight +
		v.Bus*BusWeight
}

type TrafficAnalysis struct {
	CurrentVehicleCount       int     `json:"current_vehicle_count"`
	CurrentCongestionScore    int     `json:"current_congestion_score"`
	HistoricalVehicleAverage  float64 `json:"historical_vehicle_average"`
	HistoricalCongestionAvg   float64 `json:"historical_congestion_average"`
	VehicleIncreasePercentage float64 `json:"vehicle_increase_percentage"`
	TrafficCondition          string  `json:"traffic_condition"`
}

func AnalyzeTraffic(
	ctx context.Context,
	repo *repository.TrafficRepository,
	event model.TrafficEvent,
	locationBucket string,
	timeBucket string,
	trafficDate string,
) (TrafficAnalysis, error) {

	currentVehicleCount := CalculateVehicleCount(event.Vehicles)
	currentCongestionScore := CalculateCongestionScore(event.Vehicles)

	historicalVehicleAverage, historicalCongestionAverage, err :=
		repo.GetHistoricalAverage(
			ctx,
			locationBucket,
			timeBucket,
			trafficDate,
		)

	if err != nil {
		return TrafficAnalysis{}, err
	}

	var increasePercentage float64

	if historicalVehicleAverage > 0 {
		increasePercentage =
			(float64(currentVehicleCount) - historicalVehicleAverage) /
				historicalVehicleAverage * 100
	}

	condition := "NORMAL"

	if increasePercentage >= 50 {
		condition = "VERY_HIGH"
	} else if increasePercentage >= 20 {
		condition = "HIGH"
	}

	return TrafficAnalysis{
		CurrentVehicleCount:       currentVehicleCount,
		CurrentCongestionScore:    currentCongestionScore,
		HistoricalVehicleAverage:  historicalVehicleAverage,
		HistoricalCongestionAvg:   historicalCongestionAverage,
		VehicleIncreasePercentage: increasePercentage,
		TrafficCondition:          condition,
	}, nil
}
