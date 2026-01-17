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
	"ride-hail/internal/shared/config"
	"ride-hail/internal/shared/logger"
	"ride-hail/internal/shared/postgres"
	"strconv"
)

func main() {
	logger.InitLogger("debug")
	_ = config.LoadEnv()

	getEnv := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	port := 3001
	if p, err := strconv.Atoi(getEnv("PORT", "3001")); err == nil {
		port = p
	}
	serverConfig := &handlers.ServerConfig{
		Addr: "0.0.0.0",
		Port: port,
	}

	// Read DB config from environment (with sensible defaults matching docker-compose)
	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "password")
	dbName := getEnv("POSTGRES_DB", "ride_hail")
	dbSSL := getEnv("POSTGRES_SSLMODE", "disable")

	dbConfig := postgres.NewDBConfig(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)
	slog.Info("connecting to database",
		"host", dbHost,
		"port", dbPort,
		"user", dbUser,
		"db", dbName,
		"sslmode", dbSSL,
	)
	db := postgres.NewDB(dbConfig)
	if err := db.Connect(context.Background()); err != nil {
		slog.Error("failed to connect to the database", "error", err.Error())
		os.Exit(1)
	}

	secret := getEnv("SECRET_KEY", getEnv("SECRET-KEY", "supersecretkey"))
	app := auth.NewApp([]byte(secret), serverConfig, db)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	slog.Info("starting auth service...")
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
