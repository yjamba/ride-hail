package handlers

import (
	"encoding/json"
	"net/http"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/handlers/dto"
	"ride-hail/internal/ride/service"
	"strings"
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

	// Create the ride command
	cmd := models.CreateRideCommand{
		PassengerID: req.PassengerID,
		Pickup: models.Location{
			Latitude:  req.PickupLat,
			Longitude: req.PickupLon,
			Address:   req.PickupAddr,
		},
		Destination: models.Location{
			Latitude:  req.DestLat,
			Longitude: req.DestLon,
			Address:   req.DestAddress,
		},
	}

	// Call the service to create the ride
	ride, err := h.service.CreateRide(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the created ride ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"ride_id": ride.ID,
	})
}

// CloseRide handles the cancellation of a ride
func (h *RideHandler) CloseRide(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}
	rideID := parts[len(parts)-1]

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

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Ride canceled successfully",
	})
}
