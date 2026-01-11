package handlers

import (
	"net/http"

	"ride-hail/internal/driver/handlers/middlewares"
)

func RegisterRoutes(handler *DriverHandler) http.Handler {
	mux := http.NewServeMux()

	authMiddleware := middlewares.AuthMiddleware
	middleware := middlewares.NewMiddlewareChain(middlewares.JsonMiddleware, authMiddleware)

	mux.HandleFunc("POST /drivers/{driver_id}/online", middleware.WrapHandler(handler.ChangeDriverStatusToOnline))
	// mux.HandleFunc("POST /drivers/{driver_id}/offline", nil)
	// mux.HandleFunc("POST /drivers/{driver_id}/location", nil)
	// mux.HandleFunc("POST /drivers/{driver_id}/start", nil)
	// mux.HandleFunc("POST /drivers/{driver_id}/complete", nil)

	return mux
}
