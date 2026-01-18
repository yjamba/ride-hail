package ports

import (
	"context"
	"ride-hail/internal/driver/domain/models"
)

type DriverSessionsRepository interface {
	GetById(ctx context.Context, id string) (*models.DriverSession, error)
	Create(ctx context.Context, driverId string) (string, error)
	Update(ctx context.Context, driverSession *models.DriverSession) error
	Close(ctx context.Context, driverSessionId string) error
	GetActiveByDriverID(ctx context.Context, driverID string) (*models.DriverSession, error)
}
