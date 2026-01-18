package handlers

import (
	"net/http"

	"ride-hail/internal/ride/handlers/middleware"
)

func RegisterRoutes(handler *RideHandler, secretKey []byte) http.Handler {
	mux := http.NewServeMux()

	// REST API routes
	mux.Handle("POST /rides", jsonPost(handler.CreateRide))
	mux.Handle("POST /rides/{ride_id}/cancel", jsonPost(handler.CloseRide))

	// WebSocket route for passengers
	mux.HandleFunc("GET /ws/passengers/{passenger_id}", PassengerWSHandler(secretKey))

	return mux
}

func jsonPost(h http.HandlerFunc) http.Handler {
	return middleware.JsonMiddleware(h)
}
