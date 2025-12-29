package auth

import (
	"context"
	"ride-hail/internal/auth/handlers"
	"ride-hail/internal/auth/repository"
	"ride-hail/internal/auth/service"
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
	userRepository := repository.NewUserRepository(a.db)
	driverRepository := repository.NewDriverRepository(a.db)

	authService := service.NewAuthService(userRepository, driverRepository, []byte("supersecretkey"))
	authHandler := handlers.NewAuthHandler(authService)

	a.server = handlers.NewServer(authHandler, a.config)

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
