package ports

import (
	"context"

	"ride-hail/internal/ride/domain/models"
)

type RideRepository interface {
	CreateRide(ctx context.Context, ride *models.Ride) error
	ListByPassenger(ctx context.Context, passengerID string) ([]models.Ride, error)
	ListByStatus(ctx context.Context, passengerID string, status string) ([]models.Ride, error)
	GetRide(ctx context.Context, id string) (models.Ride, error)
	UpdateStatus(ctx context.Context, rideID string, status string) error
	CloseRide(ctx context.Context, id string, reason string) error
}
