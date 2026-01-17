package ports

import (
	"context"

	"ride-hail/internal/driver/domain/models"
)

type HistoryLocationRepository interface {
	AddLocation(ctx context.Context, locationHistory *models.LocationHistory) error
}
