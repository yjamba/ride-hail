package ride

import (
	"context"

	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/ride/repository"
	"ride-hail/internal/ride/service"
	"ride-hail/internal/shared/broker/rabbitmq"
	"ride-hail/internal/shared/postgres"
)

type App struct {
	config *handlers.ServerConfig
	db     *postgres.Database
	rmq    *rabbitmq.RMQ

	server *handlers.Server
}

func NewApp(config *handlers.ServerConfig, db *postgres.Database, rmq *rabbitmq.RMQ) *App {
	return &App{
		config: config,
		db:     db,
		rmq:    rmq,
	}
}

func (a *App) Start(ctx context.Context) error {
	repository := repository.NewRideRepo(a.db)

	service := service.NewRideService(repository, []byte("supersecretkey"))
	handler := handlers.NewRideHandler(service)

	a.server = handlers.NewServer(handler, a.config)

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
