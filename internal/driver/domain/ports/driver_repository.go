package ports

import (
	"context"

	"ride-hail/internal/driver/domain/models"
)

type DriverRepository interface {
	GetById(ctx context.Context, id string) (*models.Driver, error)
	Update(ctx context.Context, driver *models.Driver) error
	UpdateStatus(ctx context.Context, id string, status models.DriverStatus) error
}
