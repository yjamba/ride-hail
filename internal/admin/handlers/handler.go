package handlers

import (
	"net/http"

	"ride-hail/internal/admin/service"
)

type Handler struct {
	service service.Service
}

func NewHandler(s service.Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (s *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
}

func (s *Handler) GetRidesList(w http.ResponseWriter, r *http.Request) {
}
