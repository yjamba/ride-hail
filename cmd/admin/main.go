package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"ride-hail/internal/admin"
	"ride-hail/internal/admin/handlers"
	"ride-hail/internal/shared/config"
	"ride-hail/internal/shared/logger"
	"ride-hail/internal/shared/postgres"
)

func main() {
	logger := logger.InitLogger("debug")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = config.LoadEnv()
	getEnv := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "password")
	dbName := getEnv("POSTGRES_DB", "ride_hail")
	dbSSL := getEnv("POSTGRES_SSLMODE", "disable")

	dbConfig := postgres.NewDBConfig(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)
	db := postgres.NewDB(dbConfig)
	if err := db.Connect(ctx); err != nil {
		slog.Error("failed to connect to postgres", "error", err.Error())
		os.Exit(1)
	}

	port := 3003
	if p, err := strconv.Atoi(getEnv("RIDE_SERVICE_PORT", "3003")); err == nil {
		port = p
	}

	config := &handlers.ServerConfig{
		Addr: getEnv("RIDE_SERVICE_HOST", "0.0.0.0"),
		Port: port,
	}

	app := admin.NewApp(db, config, logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)
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

	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			slog.Error("failed to start auth service", "error", err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	if err := app.Stop(ctx); err != nil {
		slog.Error("failed to stop auth service", "error", err.Error())
	}

	wg.Wait()
}
