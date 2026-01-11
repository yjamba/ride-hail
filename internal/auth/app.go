package auth

import (
	"context"

	"ride-hail/internal/auth/handlers"
	"ride-hail/internal/auth/repository"
	"ride-hail/internal/auth/service"
	"ride-hail/internal/shared/postgres"
)

type App struct {
	secretKey []byte

	config *handlers.ServerConfig
	db     *postgres.Database

	server *handlers.Server
}

func NewApp(secretKey []byte, config *handlers.ServerConfig, db *postgres.Database) *App {
	return &App{
		secretKey: secretKey,
		config:    config,
		db:        db,
	}
}

func (a *App) Start(ctx context.Context) error {
	userRepository := repository.NewUserRepository(a.db)
	driverRepository := repository.NewDriverRepository(a.db)
	txManager := postgres.NewTxManager(a.db)

	tokenService := service.NewTokenService(a.secretKey)
	authService := service.NewAuthService(tokenService, userRepository, driverRepository, txManager)
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

	if a.db != nil {
		return a.db.Close()
	}

	return nil
}
