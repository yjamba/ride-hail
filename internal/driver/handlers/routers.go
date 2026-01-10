package handlers

import (
	"net/http"

	"ride-hail/internal/driver/handlers/middlewares"
)

func RegisterRoutes(handler *DriverHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /drivers/{driver_id}/online", middlewares.AuthMiddleware(handler.ChangeDriverStatusToOnline))
	// mux.HandleFunc("/drivers/{driver_id}/offline", nil)
	// mux.HandleFunc("/drivers/{driver_id}/location", nil)
	// mux.HandleFunc("/drivers/{driver_id}/start", nil)
	// mux.HandleFunc("/drivers/{driver_id}/complete", nil)

	return mux
}
