package handlers

import (
	"encoding/json"
	"net/http"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/handlers/dto"
	"ride-hail/internal/ride/service"
)

type RideHandler struct {
	service service.RideService
}

func NewRideHandler(service service.RideService) *RideHandler {
	return &RideHandler{service: service}
}

type ctxKey string

const PassengerIDKey ctxKey = "passenger_id"

func (h *RideHandler) CreateRide(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req dto.CreateRideRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	passengerID, ok := r.Context().Value(PassengerIDKey).(string)
	if !ok || passengerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vehicleType := models.VehicleType(req.RideType)
	if !vehicleType.IsValid() {
		http.Error(w, "invalid vehicle type", http.StatusBadRequest)
		return
	}

	cmd := models.CreateRideCommand{
		PassengerID: passengerID,
		VehicleType: vehicleType,
		Pickup: models.Location{
			Latitude:  req.PickupLat,
			Longitude: req.PickupLon,
			Address:   req.PickupAddress,
		},
		Destination: models.Location{
			Latitude:  req.DestLat,
			Longitude: req.DestLon,
			Address:   req.DestAddress,
		},
	}

	ride, err := h.service.CreateRide(r.Context(), cmd)
	if err != nil {
		// минимальный маппинг
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"ride_id": ride.ID,
	})
}

func (h *RideHandler) CloseRide(w http.ResponseWriter, r *http.Request) {
}
