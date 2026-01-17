package admin

import (
	"context"
	"log/slog"

	"ride-hail/internal/admin/handlers"
	"ride-hail/internal/admin/repository"
	"ride-hail/internal/admin/service"
	"ride-hail/internal/shared/postgres"
)

type App struct {
	server       *handlers.Server
	serverConfig *handlers.ServerConfig

	db *postgres.Database

	logger *slog.Logger
}

func NewApp(db *postgres.Database, serverConfig *handlers.ServerConfig, logger *slog.Logger) *App {
	return &App{
		db:     db,
		logger: logger,
	}
}

func (a *App) Start(ctx context.Context) error {
	metricsRepo := repository.NewMetricsRepository(a.db)
	ridesRepo := repository.NewRidesRepository(a.db)

	service := service.NewService(metricsRepo, ridesRepo, a.logger)

	handler := handlers.NewHandler(*service)

	a.server = handlers.NewServer(handler, a.serverConfig)

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
