package driver

import (
	"context"
	"log/slog"

	"ride-hail/internal/driver/handlers"
)

type App struct {
	server *handlers.Server
}

func NewApp() *App {
	return &App{}
}

func (a *App) Start(ctx context.Context) error {
	config := handlers.NewServerConfig("0.0.0.0", 3002)
	handler := handlers.NewDriverHandler()
	a.server = handlers.NewServer(handler, config)

	if err := a.server.Start(ctx); err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	slog.Info("Stopping driver service")

	if a.server != nil {
		return a.server.Stop(ctx)
	}

	return nil
}
