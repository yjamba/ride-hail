package ride

import (
	"context"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/ride/repository"
	"ride-hail/internal/ride/service"
)

type App struct {
	config *handlers.ServerConfig
	db     *repository.DB

	server *handlers.Server
}

func NewApp(config *handlers.ServerConfig, db *repository.DB) *App {
	return &App{
		config: config,
		db:     db,
	}
}

func (a *App) Start(ctx context.Context) error {
	Repository := repository.NewRideRepo(a.db)

	Service := service.NewRideService(Repository, []byte("supersecretkey"))
	Handler := handlers.NewRideHandler(Service)

	a.server = handlers.NewServer(Handler, a.config)

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
