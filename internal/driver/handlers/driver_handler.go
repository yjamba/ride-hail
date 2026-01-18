package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/services"
)

const path_value string = "driver_id"

type DriverHandler struct {
	service *services.DriverService
}

func NewDriverHandler() *DriverHandler {
	return &DriverHandler{}
}

func (h *DriverHandler) ChangeDriverStatusToOnline(w http.ResponseWriter, r *http.Request) {
	driver_id := r.PathValue(path_value)
	defer r.Body.Close()
	var req models.Location

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	session_id, err := h.service.GoOnline(r.Context(), driver_id, req.Latitude, req.Longitude)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "AVAILABLE",
		"session_id": session_id,
		"message":    "You are now online and ready to accept rides",
	})
}

func (h *DriverHandler) ChangeDriverStatusToOffline(w http.ResponseWriter, r *http.Request) {
	driver_id := r.PathValue(path_value)
	defer r.Body.Close()

	session_summary, err := h.service.GoOffline(r.Context(), driver_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := models.OfflineResponse{
		Status:         "OFFLINE",
		SessionID:      session_summary.SessionID,
		SessionSummary: *session_summary,
		Message:        "You are now offline",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *DriverHandler) UpdateDriverLocation(w http.ResponseWriter, r *http.Request) {
	driver_id := r.PathValue(path_value)
	defer r.Body.Close()
	var req models.LocationUpdate
	var loc models.Location

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	coordID, err := h.service.UpdateLocation(r.Context(), driver_id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loc.Latitude = req.Latitude
	loc.Longitude = req.Longitude

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"coordinate_id": coordID,
		"updated_at": time.Now().Format(time.RFC3339),
	})
}

func (h *DriverHandler) StartRide(w http.ResponseWriter, r *http.Request) {
	driver_id := r.PathValue(path_value)
	defer r.Body.Close()
	var req models.StartDriveRequest

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	err := h.service.StartRide(r.Context(), driver_id, req.RideID, req.Location.Latitude, req.Location.Longitude)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"ride_id": req.RideID,
		"status":  "BUSY",
		"started_at": time.Now().Format(time.RFC3339),
		"message": "Ride started successfully",
	})
}

func (h *DriverHandler) CompleteRide(w http.ResponseWriter, r *http.Request) {
	driver_id := r.PathValue(path_value)
	defer r.Body.Close()
	var req models.CompleteRideRequest
	var earnings float64

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	earnings, err := h.service.CompleteRide(r.Context(), driver_id, req.RideID, req.Location.Latitude, req.Location.Longitude, req.ActualDistanceKm, req.ActualDurationMinutes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"ride_id": req.RideID,
		"status":  "AVAILABLE",
		"completed_at": time.Now().Format(time.RFC3339),
		"driver_earnings":  fmt.Sprintf("%.2f", earnings),
		"message": "Ride completed successfully",
	})
}
