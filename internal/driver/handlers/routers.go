package handlers

import (
	"net/http"

	"ride-hail/internal/driver/handlers/middlewares"
	"ride-hail/internal/driver/handlers/ws"
)

func RegisterRoutes(handler *DriverHandler, ws *ws.WSHandler) http.Handler {
	mux := http.NewServeMux()

	authMiddleware := middlewares.AuthMiddleware
	middleware := middlewares.NewMiddlewareChain(middlewares.JsonMiddleware, authMiddleware)

	mux.HandleFunc("POST /drivers/{driver_id}/online", middleware.WrapHandler(handler.ChangeDriverStatusToOnline))
	mux.HandleFunc("/ws/drivers/{driver_id}", middleware.WrapHandler(ws.ServeWS))
	mux.HandleFunc("POST /drivers/{driver_id}/offline", middleware.WrapHandler(handler.ChangeDriverStatusToOffline))
	mux.HandleFunc("POST /drivers/{driver_id}/location", middleware.WrapHandler(handler.UpdateDriverLocation))
	mux.HandleFunc("POST /drivers/{driver_id}/start", middleware.WrapHandler(handler.StartRide))
	// mux.HandleFunc("POST /drivers/{driver_id}/complete", nil)

	return mux
}
