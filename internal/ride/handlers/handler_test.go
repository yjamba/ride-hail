package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateRide_InvalidJSON(t *testing.T) {
	h := NewRideHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	h.CreateRide(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCloseRide_InvalidURL(t *testing.T) {
	h := NewRideHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader("{}"))
	rr := httptest.NewRecorder()

	h.CloseRide(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCloseRide_InvalidJSON(t *testing.T) {
	h := NewRideHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/rides/123", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	h.CloseRide(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}