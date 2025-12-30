package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ride-hail/internal/auth"
	"ride-hail/internal/auth/handlers"
	"ride-hail/internal/shared/logger"
	"ride-hail/internal/shared/postgres"
)

func main() {
	logger.InitLogger("debug")

	config := &handlers.ServerConfig{
		Addr: "localhost",
		Port: 3001,
	}

	dbConfig := postgres.NewDBConfig("localhost", "5432", "postgres", "postgres", "ride-hail", "disabled")
	db := postgres.NewDB(dbConfig)

	app := auth.NewApp(config, db)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := app.Start(context.Background()); err != nil {
			slog.Error("failed to start auth service", "error", err.Error())
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for range ticker.C {
			if !db.IsHealthy(context.Background()) {
				slog.Error("database connection is not healthy")
				if err := db.Reconnect(context.Background()); err != nil {
					slog.Error("failed to reconnect to the database", "error", err.Error())
				} else {
					slog.Info("successfully reconnected to the database")
				}
			}
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
