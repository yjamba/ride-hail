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

func (h *RideHandler) CreateRide(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRideRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	passengerID := r.Context().Value("passenger_id").(string)
	cmd := models.CreateRideCommand{
		PassengerID: passengerID,
		RideType:    req.RideType,
	}

	cmd.Pickup = models.Location{
		Lat:     req.PickupLat,
		Lon:     req.PickupLon,
		Address: req.PickupAddress,
	}

	cmd.Destination = models.Location{
		Lat:     req.DestLat,
		Lon:     req.DestLon,
		Address: req.DestAddress,
	}
}

func (h *RideHandler) CloseRide(w http.ResponseWriter, r *http.Request) {
}
