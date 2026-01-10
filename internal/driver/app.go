package driver

import "context"

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Start(ctx context.Context) error {
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	return nil
}
