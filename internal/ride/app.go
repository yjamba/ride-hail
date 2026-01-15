package ride

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
	Repository := repository.NewDriverRepository(a.db)

	Service := service.NewAuthService(userRepository, Repository, []byte("supersecretkey"))
	Handler := handlers.NewAuthHandler(Service)

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
