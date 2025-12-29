package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"ride-hail/internal/auth"
	"ride-hail/internal/auth/handlers"
	"ride-hail/internal/auth/repository"
	"ride-hail/internal/shared/logger"
	"sync"
	"syscall"
)

func main() {
	logger.InitLogger("debug")

	config := &handlers.ServerConfig{
		Addr: "localhost",
		Port: 3001,
	}

	db := &repository.DB{}

	app := auth.NewApp(config, db)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := app.Start(context.Background()); err != nil {
			slog.Error("failed to start auth service", "error", err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Stop(ctx); err != nil {
		slog.Error("failed to stop auth service", "error", err.Error())
	}

	wg.Wait()
}
