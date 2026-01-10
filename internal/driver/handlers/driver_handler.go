package handlers

import "net/http"

type DriverHandler struct{}

func NewDriverHandler() *DriverHandler {
	return &DriverHandler{}
}

func (h *DriverHandler) ChangeDriverStatusToOnline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Hello, World"))
}
