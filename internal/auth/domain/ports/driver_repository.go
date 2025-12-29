package ports

import (
	"context"

	"ride-hail/internal/auth/domain/models"
)

type DriverRepository interface {
	Save(ctx context.Context, driver *models.Driver) (string, error)
}
