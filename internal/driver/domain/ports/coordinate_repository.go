package ports

import (
	"context"
	"ride-hail/internal/driver/domain/models"
)

type CoordinateRepository interface {
	CreateOrUpdate(ctx context.Context, entityID, entityType string,
		lat, lon float64, address string,
	) (string, error)
	GetCurrent(ctx context.Context, entityID, entityType string) (*models.Coordinate, error)
}
