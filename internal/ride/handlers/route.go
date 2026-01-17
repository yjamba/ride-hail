package handlers

import (
	"net/http"
	"ride-hail/internal/ride/handlers/middleware"
)

func RegisterRoutes(handler *RideHandler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /rides", jsonPost(handler.CreateRide))
	mux.Handle("POST /rides/cancel", jsonPost(handler.CloseRide))

	return mux
}

func jsonPost(h http.HandlerFunc) http.Handler {
	return middleware.JsonMiddleware(h)
}
