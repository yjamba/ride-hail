package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/handlers/dto"
	"ride-hail/internal/ride/service"
)

type RideHandler struct {
	service *service.RideService
}

func NewRideHandler(service *service.RideService) *RideHandler {
	return &RideHandler{service: service}
}

// CreateRide handles the creation of a new ride
func (h *RideHandler) CreateRide(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req dto.CreateRideRequest

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Default to ECONOMY if not specified
	vehicleType := models.VehicleType(req.RideType)
	if !vehicleType.IsValid() {
		vehicleType = models.VehicleTypeEconomy
	}

	// Create the ride command
	cmd := models.CreateRideCommand{
		PassengerID: req.PassengerID,
		VehicleType: vehicleType,
		Pickup: models.Location{
			Latitude:  req.PickupLatitude,
			Longitude: req.PickupLongitude,
			Address:   req.PickupAddress,
		},
		Destination: models.Location{
			Latitude:  req.DestinationLatitude,
			Longitude: req.DestinationLongitude,
			Address:   req.DestinationAddress,
		},
	}

	// Call the service to create the ride
	ride, err := h.service.CreateRide(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build response per README spec
	resp := dto.RideResponse{
		RideID:                   ride.ID,
		RideNumber:               ride.RideNumber,
		Status:                   string(ride.Status),
		EstimatedFare:            getFloat(ride.EstimatedFare),
		EstimatedDurationMinutes: ride.EstimatedDurationMinutes,
		EstimatedDistanceKm:      ride.EstimatedDistanceKm,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// CloseRide handles the cancellation of a ride
func (h *RideHandler) CloseRide(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	rideID := r.PathValue("ride_id")
	if rideID == "" {
		http.Error(w, "ride_id is required", http.StatusBadRequest)
		return
	}

	// Decode the JSON request body
	var req dto.CancelRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Call the service to close the ride
	if err := h.service.CloseRide(r.Context(), rideID, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond per README spec
	resp := dto.CancelRideResponse{
		RideID:      rideID,
		Status:      "CANCELLED",
		CancelledAt: time.Now().Format(time.RFC3339),
		Message:     "Ride cancelled successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func getFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
