package admin

import (
	"context"

	"ride-hail/internal/admin/handlers"
	"ride-hail/internal/admin/repository"
	"ride-hail/internal/admin/service"
	"ride-hail/internal/shared/logger"
	"ride-hail/internal/shared/postgres"
)

type App struct {
	server       *handlers.Server
	serverConfig *handlers.ServerConfig

	db *postgres.Database

	logger *logger.Logger
}

func NewApp(db *postgres.Database, serverConfig *handlers.ServerConfig, log *logger.Logger) *App {
	return &App{
		db:           db,
		serverConfig: serverConfig,
		logger:       log,
	}
}

func (a *App) Start(ctx context.Context) error {
	metricsRepo := repository.NewMetricsRepository(a.db)
	ridesRepo := repository.NewRidesRepository(a.db)

	svc := service.NewService(metricsRepo, ridesRepo, a.logger)

	handler := handlers.NewHandler(*svc)

	a.server = handlers.NewServer(handler, a.serverConfig)

	if a.logger != nil {
		a.logger.Info(ctx, "server_starting", "Admin service starting")
	}

	if err := a.server.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.server != nil {
		if err := a.server.Stop(ctx); err != nil {
			return err
		}

		if err := a.db.Close(); err != nil {
			return err
		}
	}
	return nil
}
