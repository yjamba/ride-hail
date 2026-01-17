package handlers

import (
	"net/http"

	"ride-hail/internal/admin/handlers/middleware"
)

func RegisterRoutes(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /admin/overview", middleware.AuthMiddleware(handler.GetOverview))
	mux.HandleFunc("GET /admin/rides/active", middleware.AuthMiddleware(handler.GetRidesList))

	return mux
}
