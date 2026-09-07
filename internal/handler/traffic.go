package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"traffic-backend/internal/model"
	"traffic-backend/internal/repository"
	"traffic-backend/internal/service"
)

type TrafficHandler struct {
	repo *repository.TrafficRepository
}

func NewTrafficHandler(repo *repository.TrafficRepository) *TrafficHandler {
	return &TrafficHandler{
		repo: repo,
	}
}

// HandleTrafficEvent receives a traffic event from the AI/event backend.
func (h *TrafficHandler) HandleTrafficEvent(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Decode incoming traffic event JSON.
	var event model.TrafficEvent

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&event); err != nil {
		http.Error(
			w,
			"invalid JSON: "+err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	// Calculate total vehicle count from individual vehicle classes.
	vehicleCount := service.CalculateVehicleCount(
		event.Vehicles,
	)

	// Calculate weighted congestion score.
	congestionScore := service.CalculateCongestionScore(
		event.Vehicles,
	)

	// Create geographic bucket.
	locationBucket := service.CreateLocationBucket(
		event.Location,
	)

	// Create 5-minute time bucket.
	timeBucket := service.CreateTimeBucket(
		event.Timestamp,
	)

	// Extract traffic date.
	trafficDate := event.Timestamp.Format("2006-01-02")

	// Store raw event and update geographic/time aggregate
	// inside a single database transaction.
	err := h.repo.SaveEventAndUpdateAggregate(
		r.Context(),
		event,
		locationBucket,
		timeBucket,
		trafficDate,
		vehicleCount,
		congestionScore,
		event.Traffic.VehicleOccupancy,
	)

	if err != nil {

		// Duplicate event protection.
		if err.Error() == "duplicate event_id" {
			http.Error(
				w,
				"duplicate event_id",
				http.StatusConflict,
			)
			return
		}

		log.Println(
			"SaveEventAndUpdateAggregate ERROR:",
			err,
		)

		http.Error(
			w,
			"failed to process traffic event: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	// Event has been successfully processed.
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	response := map[string]interface{}{
		"success":  true,
		"event_id": event.EventID,
		"message":  "traffic event processed successfully",
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// HandleCurrentTraffic returns the current geographic traffic
// aggregates in the format expected by the frontend.
func (h *TrafficHandler) HandleCurrentTraffic(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Get today's geographic traffic aggregates.
	aggregates, err := h.repo.GetCurrentTraffic(
		r.Context(),
	)

	if err != nil {

		log.Println(
			"GetCurrentTraffic ERROR:",
			err,
		)

		http.Error(
			w,
			"failed to get current traffic: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	// Convert database aggregates into frontend JSON.
	responses := make(
		[]model.FrontendTrafficResponse,
		0,
		len(aggregates),
	)

	for _, aggregate := range aggregates {

		trafficDate := aggregate.TrafficDate.Format(
			"2006-01-02",
		)

		// Get historical average for the same
		// location and 5-minute time bucket.
		historicalVehicleAverage, _, err :=
			h.repo.GetHistoricalAverage(
				r.Context(),
				aggregate.LocationBucket,
				aggregate.TimeBucket,
				trafficDate,
			)

		if err != nil {

			log.Println(
				"GetHistoricalAverage ERROR:",
				err,
			)

			http.Error(
				w,
				"failed to get historical traffic: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		response :=
			service.BuildFrontendTrafficFromAggregate(
				aggregate,
				historicalVehicleAverage,
			)

		responses = append(
			responses,
			response,
		)
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(responses)
}