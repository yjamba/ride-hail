package ports

import (
	"context"

	"ride-hail/internal/admin/domain/models"
)

type MetricsRepository interface {
	FetchOverview(ctx context.Context) (*models.Overview, error)
}
