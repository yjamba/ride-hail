package service

import (
	"context"
	"errors"
	"testing"

	"ride-hail/internal/ride/domain/models"
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

func TestValidateLanLon(t *testing.T) {
	cases := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{name: "valid", lat: 43.238949, lon: 76.889709, wantErr: false},
		{name: "lat too low", lat: -91, lon: 0, wantErr: true},
		{name: "lat too high", lat: 91, lon: 0, wantErr: true},
		{name: "lon too low", lat: 0, lon: -181, wantErr: true},
		{name: "lon too high", lat: 0, lon: 181, wantErr: true},
		{name: "boundary lat min", lat: -90, lon: 0, wantErr: false},
		{name: "boundary lat max", lat: 90, lon: 0, wantErr: false},
		{name: "boundary lon min", lat: 0, lon: -180, wantErr: false},
		{name: "boundary lon max", lat: 0, lon: 180, wantErr: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateLanLon(tc.lat, tc.lon)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"REQUESTED", "IN_PROGRESS", "COMPLETED"}

	if !contains(slice, "REQUESTED") {
		t.Error("expected true for REQUESTED")
	}
	if !contains(slice, "COMPLETED") {
		t.Error("expected true for COMPLETED")
	}
	if contains(slice, "INVALID") {
		t.Error("expected false for INVALID")
	}
	if contains(nil, "any") {
		t.Error("expected false for nil slice")
	}
}

func TestNewRideService(t *testing.T) {
	repo := &mockRideRepo{}
	secret := []byte("test-secret")

	svc := NewRideService(repo, nil, nil, secret)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.repo != repo {
		t.Error("repo not set correctly")
	}
}

func TestCreateRide_Success(t *testing.T) {
	repo := &mockRideRepo{}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	cmd := models.CreateRideCommand{
		PassengerID: "passenger-123",
		Pickup:      models.Location{Latitude: 43.238949, Longitude: 76.889709, Address: "Pickup"},
		Destination: models.Location{Latitude: 43.222015, Longitude: 76.851511, Address: "Dest"},
	}

	ride, err := svc.CreateRide(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ride == nil {
		t.Fatal("expected ride, got nil")
	}
	if ride.Status != models.RideStatusRequested {
		t.Errorf("expected status %s, got %s", models.RideStatusRequested, ride.Status)
	}
}

func TestCreateRide_InvalidPickupCoords(t *testing.T) {
	repo := &mockRideRepo{}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	cmd := models.CreateRideCommand{
		PassengerID: "passenger-123",
		Pickup:      models.Location{Latitude: 100, Longitude: 76.889709}, // invalid lat
		Destination: models.Location{Latitude: 43.222015, Longitude: 76.851511},
	}

	_, err := svc.CreateRide(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid pickup coordinates")
	}
}

func TestCreateRide_InvalidDestCoords(t *testing.T) {
	repo := &mockRideRepo{}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	cmd := models.CreateRideCommand{
		PassengerID: "passenger-123",
		Pickup:      models.Location{Latitude: 43.238949, Longitude: 76.889709},
		Destination: models.Location{Latitude: 43.222015, Longitude: 200}, // invalid lon
	}

	_, err := svc.CreateRide(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid destination coordinates")
	}
}

func TestCreateRide_RepoError(t *testing.T) {
	repo := &mockRideRepo{
		createRideFunc: func(ctx context.Context, ride *models.Ride) error {
			return errors.New("db error")
		},
	}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	cmd := models.CreateRideCommand{
		PassengerID: "passenger-123",
		Pickup:      models.Location{Latitude: 43.238949, Longitude: 76.889709},
		Destination: models.Location{Latitude: 43.222015, Longitude: 76.851511},
	}

	_, err := svc.CreateRide(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetRideById_Success(t *testing.T) {
	repo := &mockRideRepo{
		getRideFunc: func(ctx context.Context, id string) (models.Ride, error) {
			return models.Ride{ID: id, PassengerID: "passenger-123"}, nil
		},
	}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	ride, err := svc.GetRideById(context.Background(), "ride-123", "passenger-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ride.ID != "ride-123" {
		t.Errorf("expected ride ID ride-123, got %s", ride.ID)
	}
}

func TestGetRideById_NotFound(t *testing.T) {
	repo := &mockRideRepo{
		getRideFunc: func(ctx context.Context, id string) (models.Ride, error) {
			return models.Ride{}, errors.New("not found")
		},
	}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	_, err := svc.GetRideById(context.Background(), "nonexistent", "passenger-123")
	if err == nil {
		t.Fatal("expected error for nonexistent ride")
	}
}

func TestGetRideByStatus_Success(t *testing.T) {
	repo := &mockRideRepo{
		listByStatusFunc: func(ctx context.Context, passengerID, status string) ([]models.Ride, error) {
			return []models.Ride{{ID: "ride-1"}, {ID: "ride-2"}}, nil
		},
	}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	rides, err := svc.GetRideByStatus(context.Background(), "passenger-123", "REQUESTED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rides) != 2 {
		t.Errorf("expected 2 rides, got %d", len(rides))
	}
}

func TestUpdateRideStatus_ValidStatus(t *testing.T) {
	repo := &mockRideRepo{}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	validStatuses := []string{"REQUESTED", "IN_PROGRESS", "COMPLETED", "CANCELLED"}
	for _, status := range validStatuses {
		err := svc.UpdateRideStatus(context.Background(), "ride-123", status)
		if err != nil {
			t.Errorf("unexpected error for status %s: %v", status, err)
		}
	}
}

func TestUpdateRideStatus_InvalidStatus(t *testing.T) {
	repo := &mockRideRepo{}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	err := svc.UpdateRideStatus(context.Background(), "ride-123", "INVALID_STATUS")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestUpdateRideStatus_RepoError(t *testing.T) {
	repo := &mockRideRepo{
		updateStatusFunc: func(ctx context.Context, rideID, status string) error {
			return errors.New("db error")
		},
	}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	err := svc.UpdateRideStatus(context.Background(), "ride-123", "COMPLETED")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestCloseRide_Success(t *testing.T) {
	repo := &mockRideRepo{}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	err := svc.CloseRide(context.Background(), "ride-123", "changed my mind")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloseRide_RepoError(t *testing.T) {
	repo := &mockRideRepo{
		closeRideFunc: func(ctx context.Context, id, reason string) error {
			return errors.New("db error")
		},
	}
	svc := NewRideService(repo, nil, nil, []byte("secret"))

	err := svc.CloseRide(context.Background(), "ride-123", "reason")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
