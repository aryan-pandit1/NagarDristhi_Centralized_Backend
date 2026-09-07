package service

import (
	"fmt"
	"math"
	"time"

	"traffic-backend/internal/model"
)

func GetCongestionLevel(score float64) string {
	switch {
	case score <= 30:
		return "LOW"
	case score <= 60:
		return "MEDIUM"
	default:
		return "HIGH"
	}
}

func BuildFrontendTrafficResponse(
	eventID string,
	timestamp time.Time,
	latitude float64,
	longitude float64,
	area string,
	city string,
	vehicleCount int,
	historicalVehicleAverage float64,
	vehicleIncreasePercentage float64,
	vehicleOccupancy float64,
	congestionLevel string,
) model.FrontendTrafficResponse {

	return model.FrontendTrafficResponse{
		EventID:   eventID,
		Timestamp: timestamp,
		Location: model.FrontendLocation{
			Latitude:  latitude,
			Longitude: longitude,
			Area:      area,
			City:      city,
		},
		Traffic: model.FrontendTraffic{
			VehicleCount:              vehicleCount,
			HistoricalVehicleAverage:  historicalVehicleAverage,
			VehicleIncreasePercentage: vehicleIncreasePercentage,
			VehicleOccupancy:          vehicleOccupancy,
			CongestionLevel:           congestionLevel,
		},
	}
}

func BuildFrontendTrafficFromAggregate(
	aggregate model.TrafficAggregate,
	historicalVehicleAverage float64,
) model.FrontendTrafficResponse {

	vehicleCount := int(math.Round(aggregate.AverageVehicleCount))

	var increasePercentage float64

	if historicalVehicleAverage > 0 {
		increasePercentage =
			(aggregate.AverageVehicleCount - historicalVehicleAverage) /
				historicalVehicleAverage * 100
	}

	congestionLevel := GetCongestionLevel(
		aggregate.AverageCongestionScore,
	)

	eventID := fmt.Sprintf(
		"%s_%s_%s",
		aggregate.LocationBucket,
		aggregate.TrafficDate.Format("2006-01-02"),
		aggregate.TimeBucket,
	)

	return BuildFrontendTrafficResponse(
		eventID,
		aggregate.UpdatedAt,
		aggregate.Latitude,
		aggregate.Longitude,
		"lalgang",
		"kanpur",
		vehicleCount,
		historicalVehicleAverage,
		increasePercentage,
		aggregate.AverageOccupancy,
		congestionLevel,
	)
}
