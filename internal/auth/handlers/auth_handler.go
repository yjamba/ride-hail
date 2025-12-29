package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/handlers/dto"
	"ride-hail/internal/auth/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) HandleSingupPassenger(w http.ResponseWriter, r *http.Request) {
	request, err := h.decodeUserRequest(r)
	if err != nil {
		slog.Error("failed to decode user request", "error", err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id, tokenPairs, err := h.authService.Signup(r.Context(), request.Email, request.Password, "PASSENGER")
	if err != nil {
		slog.Error("failed to signup passenger", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendResponse(w, id, tokenPairs)
}

func (h *AuthHandler) HandleSingupAdmin(w http.ResponseWriter, r *http.Request) {
	request, err := h.decodeUserRequest(r)
	if err != nil {
		slog.Error("failed to decode user request", "error", err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id, tokenPairs, err := h.authService.Signup(r.Context(), request.Email, request.Password, "ADMIN")
	if err != nil {
		slog.Error("failed to signup admin", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendResponse(w, id, tokenPairs)
}

func (h *AuthHandler) HandleSingupDriver(w http.ResponseWriter, r *http.Request) {
	var request dto.DriverRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		slog.Error("failed to decode driver request", "error", err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	driver := &models.Driver{
		LicenseNumber: request.LicenseNumber,
		VehicleType:   request.VehicleType,
		VehicleAttrs: &models.VehicleAttrs{
			VehicleMake:  request.VehicleAttrs.VehicleMake,
			VehicleModel: request.VehicleAttrs.VehicleModel,
			VehicleColor: request.VehicleAttrs.VehicleColor,
			VehiclePlate: request.VehicleAttrs.VehiclePlate,
			VehicleYear:  request.VehicleAttrs.VehicleYear,
		},
	}

	id, tokenPairs, err := h.authService.SignupDriver(r.Context(), request.Email, request.Password, driver)
	if err != nil {
		slog.Error("failed to signup driver", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.sendResponse(w, id, tokenPairs)
}

func (h *AuthHandler) decodeUserRequest(r *http.Request) (*dto.UserRequest, error) {
	var request dto.UserRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return nil, err
	}

	return &request, nil
}

func (h *AuthHandler) sendResponse(w http.ResponseWriter, id string, tokenPairs *models.TokenPair) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    tokenPairs.RefreshToken,
		HttpOnly: true,
	}

	response := dto.Response{
		ID:           id,
		AccessToken:  tokenPairs.AccessToken,
		RefreshToken: tokenPairs.RefreshToken,
	}

	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
