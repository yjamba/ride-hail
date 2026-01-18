package ports

import (
	"context"

	"ride-hail/internal/admin/domain/models"
)

type RidesRepository interface {
	FetchRidesList(ctx context.Context, page, pageSize int) (*models.RidesList, error)
}
