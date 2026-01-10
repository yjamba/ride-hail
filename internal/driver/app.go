package driver

import (
	"context"

	"ride-hail/internal/driver/handlers"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Start(ctx context.Context) error {
	config := handlers.NewServerConfig("localhost", 3002)
	handler := handlers.NewDriverHandler()
	server := handlers.NewServer(handler, config)

	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	return nil
}
