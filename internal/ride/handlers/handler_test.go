package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/service"
	"strings"
	"testing"
)

// Mock repository for testing
type mockRideRepo struct {
	createRideFunc   func(ctx context.Context, ride *models.Ride) error
	getRideFunc      func(ctx context.Context, id string) (models.Ride, error)
	listByStatusFunc func(ctx context.Context, passengerID, status string) ([]models.Ride, error)
	updateStatusFunc func(ctx context.Context, rideID, status string) error
	closeRideFunc    func(ctx context.Context, id, reason string) error
}

func (m *mockRideRepo) CreateRide(ctx context.Context, ride *models.Ride) error {
	if m.createRideFunc != nil {
		return m.createRideFunc(ctx, ride)
	}
	ride.ID = "test-ride-id"
	return nil
}

func (m *mockRideRepo) GetRide(ctx context.Context, id string) (models.Ride, error) {
	if m.getRideFunc != nil {
		return m.getRideFunc(ctx, id)
	}
	return models.Ride{ID: id}, nil
}

func (m *mockRideRepo) ListByPassenger(ctx context.Context, passengerID string) ([]models.Ride, error) {
	return nil, nil
}

func (m *mockRideRepo) ListByStatus(ctx context.Context, passengerID, status string) ([]models.Ride, error) {
	if m.listByStatusFunc != nil {
		return m.listByStatusFunc(ctx, passengerID, status)
	}
	return []models.Ride{}, nil
}

func (m *mockRideRepo) UpdateStatus(ctx context.Context, rideID, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, rideID, status)
	}
	return nil
}

func (m *mockRideRepo) CloseRide(ctx context.Context, id, reason string) error {
	if m.closeRideFunc != nil {
		return m.closeRideFunc(ctx, id, reason)
	}
	return nil
}

func TestNewRideHandler(t *testing.T) {
	svc := service.NewRideService(&mockRideRepo{}, []byte("secret"))
	h := NewRideHandler(svc)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCreateRide_InvalidJSON(t *testing.T) {
	h := NewRideHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	h.CreateRide(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreateRide_Success(t *testing.T) {
	repo := &mockRideRepo{}
	svc := service.NewRideService(repo, []byte("secret"))
	h := NewRideHandler(svc)

	body := `{
		"passenger_id": "passenger-123",
		"pickup_lat": 43.238949,
		"pickup_lon": 76.889709,
		"pickup_address": "Almaty Central Park",
		"dest_lat": 43.222015,
		"dest_lon": 76.851511,
		"dest_address": "Kok-Tobe Hill"
	}`

	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateRide(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestCreateRide_InvalidCoordinates(t *testing.T) {
	repo := &mockRideRepo{}
	svc := service.NewRideService(repo, []byte("secret"))
	h := NewRideHandler(svc)

	body := `{
		"passenger_id": "passenger-123",
		"pickup_lat": 100,
		"pickup_lon": 76.889709
	}`

	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateRide(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreateRide_ServiceError(t *testing.T) {
	repo := &mockRideRepo{
		createRideFunc: func(ctx context.Context, ride *models.Ride) error {
			return errors.New("db error")
		},
	}
	svc := service.NewRideService(repo, []byte("secret"))
	h := NewRideHandler(svc)

	body := `{
		"passenger_id": "passenger-123",
		"pickup_lat": 43.238949,
		"pickup_lon": 76.889709,
		"dest_lat": 43.222015,
		"dest_lon": 76.851511
	}`

	req := httptest.NewRequest(http.MethodPost, "/rides", strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/rides/123/cancel", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	h.CloseRide(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCloseRide_Success(t *testing.T) {
	repo := &mockRideRepo{}
	svc := service.NewRideService(repo, []byte("secret"))
	h := NewRideHandler(svc)

	body := `{"reason": "changed my mind"}`
	req := httptest.NewRequest(http.MethodPost, "/rides/ride-123/cancel", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CloseRide(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestCloseRide_ServiceError(t *testing.T) {
	repo := &mockRideRepo{
		closeRideFunc: func(ctx context.Context, id, reason string) error {
			return errors.New("db error")
		},
	}
	svc := service.NewRideService(repo, []byte("secret"))
	h := NewRideHandler(svc)

	body := `{"reason": "changed my mind"}`
	req := httptest.NewRequest(http.MethodPost, "/rides/ride-123/cancel", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CloseRide(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}