package service

import (
	"context"

	"ride-hail/internal/admin/domain/models"
	"ride-hail/internal/admin/domain/ports"
	"ride-hail/internal/shared/logger"
)

type Service struct {
	metricsRepo ports.MetricsRepository
	ridesRepo   ports.RidesRepository
	logger      *logger.Logger
}

func NewService(metrics ports.MetricsRepository, rides ports.RidesRepository, log *logger.Logger) *Service {
	return &Service{
		metricsRepo: metrics,
		ridesRepo:   rides,
		logger:      log,
	}
}

func (s *Service) CollectRuntimeMetrics(ctx context.Context) (*models.Overview, error) {
	return s.metricsRepo.FetchOverview(ctx)
}

func (s *Service) CollectRidesInfo(ctx context.Context, page, pageSize int) (*models.RidesList, error) {
	return s.ridesRepo.FetchRidesList(ctx, page, pageSize)
}
