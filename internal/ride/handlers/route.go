package handlers

import (
	"net/http"

	"ride-hail/internal/ride/handlers/middleware"
)

func RegisterRoutes(handler *RideHandler, secretKey []byte) http.Handler {
	mux := http.NewServeMux()

	// REST API routes with passenger authentication
	mux.Handle("POST /rides", middleware.JsonMiddleware(middleware.PassengerAuthMiddleware(handler.CreateRide)))
	mux.Handle("POST /rides/{ride_id}/cancel", middleware.JsonMiddleware(middleware.PassengerAuthMiddleware(handler.CloseRide)))

	// WebSocket route for passengers
	mux.HandleFunc("GET /ws/passengers/{passenger_id}", PassengerWSHandler(secretKey))

	return mux
}
