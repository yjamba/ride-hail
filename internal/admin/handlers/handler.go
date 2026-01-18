package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	result, err := s.service.CollectRuntimeMetrics(r.Context())
	if err != nil {
		http.Error(w, "Failed to collect runtime metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (s *Handler) GetRidesList(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		http.Error(w, "Invalid page_size parameter", http.StatusBadRequest)
		return
	}

	result, err := s.service.CollectRidesInfo(r.Context(), page, pageSize)
	if err != nil {
		http.Error(w, "Failed to get rides list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
