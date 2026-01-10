package handlers

import "net/http"

func RegisterRoutes(handler DriverHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/drivers/{driver_id}/online", nil)
	mux.HandleFunc("/drivers/{driver_id}/offline", nil)
	mux.HandleFunc("/drivers/{driver_id}/location", nil)
	mux.HandleFunc("/drivers/{driver_id}/start", nil)
	mux.HandleFunc("/drivers/{driver_id}/online", nil)
	mux.HandleFunc("/drivers/{driver_id}/complete", nil)

	return mux
}
