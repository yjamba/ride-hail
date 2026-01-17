package ports

import (
	"context"

	"ride-hail/internal/driver/domain/models"
)

type DriverSessionsRepository interface {
	Create(ctx context.Context, driverId string) error
	Update(ctx context.Context, driverSession *models.DriverSession) error
	Close(ctx context.Context, driverSessionId string) error
}
