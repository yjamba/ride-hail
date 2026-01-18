package ride

import (
	"context"
	"ride-hail/internal/ride/domain/ports"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/ride/repository"
	"ride-hail/internal/ride/service"
	"ride-hail/internal/shared/logger"
)

type App struct {
	config    *handlers.ServerConfig
	db        *repository.DB
	publisher ports.Publish
	logger    *logger.Logger

	server *handlers.Server
}

func NewApp(config *handlers.ServerConfig, db *repository.DB, publisher ports.Publish, log *logger.Logger) *App {
	return &App{
		config:    config,
		db:        db,
		publisher: publisher,
		logger:    log,
	}
}

func (a *App) Start(ctx context.Context) error {
	Repository := repository.NewRideRepo(a.db)

	Service := service.NewRideService(Repository, a.publisher, a.logger, []byte("supersecretkey"))
	Handler := handlers.NewRideHandler(Service)

	a.server = handlers.NewServer(Handler, a.config)

	a.logger.Info(ctx, "server_starting", "Ride service starting")

	if err := a.server.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.server != nil {
		return a.server.Stop(ctx)
	}

	return nil
}
