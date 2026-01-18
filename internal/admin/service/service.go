package service

import (
	"context"
	"log/slog"

	"ride-hail/internal/admin/domain/models"
	"ride-hail/internal/admin/domain/ports"
)

type Service struct {
	metricsRepo ports.MetricsRepository
	ridesRepo   ports.RidesRepository
	logger      *slog.Logger
}

func NewService(metrics ports.MetricsRepository, rides ports.RidesRepository, logger *slog.Logger) *Service {
	return &Service{
		metricsRepo: metrics,
		ridesRepo:   rides,
		logger:      logger,
	}
}

func (s *Service) CollectRuntimeMetrics(ctx context.Context) (*models.Overview, error) {
	return s.metricsRepo.FetchOverview(ctx)
}

func (s *Service) CollectRidesInfo(ctx context.Context, page, pageSize int) (*models.RidesList, error) {
	return s.ridesRepo.FetchRidesList(ctx, page, pageSize)
}
