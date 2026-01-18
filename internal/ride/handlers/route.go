package handlers

import (
	"net/http"

	"ride-hail/internal/driver/handlers/middlewares"
	"ride-hail/internal/ride/handlers/middleware"
)

func RegisterRoutes(handler *RideHandler, secretKey []byte) http.Handler {
	mux := http.NewServeMux()

	authMiddleware := middlewares.AuthMiddleware
	middleware := middlewares.NewMiddlewareChain(middlewares.JsonMiddleware, authMiddleware)

	// REST API routes
	mux.Handle("POST /rides", middleware.WrapHandler(handler.CreateRide))
	mux.Handle("POST /rides/{ride_id}/cancel", middleware.WrapHandler(handler.CloseRide))

	// WebSocket route for passengers
	mux.HandleFunc("GET /ws/passengers/{passenger_id}", middleware.WrapHandler(PassengerWSHandler(secretKey)))

	return mux
}

func jsonPost(h http.HandlerFunc) http.Handler {
	return middleware.JsonMiddleware(h)
}
