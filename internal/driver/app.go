package driver

import (
	"context"
	"log/slog"

	"ride-hail/internal/driver/handlers"
	"ride-hail/internal/driver/handlers/ws"
	"ride-hail/internal/driver/repositories"
	"ride-hail/internal/driver/services"
	"ride-hail/internal/shared/broker/rabbitmq"
	"ride-hail/internal/shared/postgres"
)

type App struct {
	server *handlers.Server
	db     *postgres.Database
	rmq    *rabbitmq.RMQ
	hub    *ws.Hub
}

func NewApp(db *postgres.Database, rmq *rabbitmq.RMQ) *App {
	return &App{
		db:  db,
		rmq: rmq,
		hub: ws.NewHub(),
	}
}

func (a *App) Start(ctx context.Context) error {
	// Initialize repositories
	driverRepo := repositories.NewDriverRepository(a.db)
	sessionRepo := repositories.NewDriverSessionsRepository(a.db)
	locationRepo := repositories.NewHistoryLocationRepository(a.db)
	coordinateRepo := repositories.NewCoordinateRepository(a.db)
	txManager := postgres.NewTxManager(a.db)

	// Initialize service
	driverService := services.NewDriverService(
		driverRepo,
		sessionRepo,
		locationRepo,
		coordinateRepo,
		a.rmq, // consume
		a.rmq, // publish
		txManager,
	)

	// Initialize handlers
	handler := handlers.NewDriverHandler(driverService)
	wsHandler := ws.NewWSHandler(a.hub)

	// Start WebSocket hub
	a.hub.Start()

	// Initialize and start server
	config := handlers.NewServerConfig("0.0.0.0", 3002)
	a.server = handlers.NewServer(handler, wsHandler, config)

	if err := a.server.Start(ctx); err != nil {
		slog.Error("failed to start server", "error", err.Error())
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	slog.Info("Stopping driver service")

	// Stop WebSocket hub
	if a.hub != nil {
		a.hub.Stop()
	}

	// Stop server
	if a.server != nil {
		if err := a.server.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
