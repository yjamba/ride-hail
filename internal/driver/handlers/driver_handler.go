package handlers

import "net/http"

const path_value string = "driver_id"

type DriverHandler struct{}

func NewDriverHandler() *DriverHandler {
	return &DriverHandler{}
}

func (h *DriverHandler) ChangeDriverStatusToOnline(w http.ResponseWriter, r *http.Request) {
	driver_id := r.PathValue(path_value)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(driver_id))
}
