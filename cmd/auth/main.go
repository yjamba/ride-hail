package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"ride-hail/internal/auth"
	"ride-hail/internal/auth/handlers"
	"ride-hail/internal/shared/config"
	"ride-hail/internal/shared/logger"
	"ride-hail/internal/shared/postgres"
)

func main() {
	_ = config.LoadEnv()

	getEnv := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	// Initialize structured logger
	log := logger.NewLogger("auth-service", getEnv("LOG_LEVEL", "info"))
	ctx := context.Background()

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
	log.InfoWithFields(ctx, "db_connecting", "Connecting to database", map[string]interface{}{
		"host":    dbHost,
		"port":    dbPort,
		"user":    dbUser,
		"db":      dbName,
		"sslmode": dbSSL,
	})

	db := postgres.NewDB(dbConfig)
	if err := db.Connect(ctx); err != nil {
		log.Error(ctx, "db_connect_error", "Failed to connect to the database", err)
		os.Exit(1)
	}
	log.Info(ctx, "db_connected", "Successfully connected to database")

	secret := getEnv("SECRET_KEY", getEnv("SECRET-KEY", "supersecretkey"))
	app := auth.NewApp([]byte(secret), serverConfig, db)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	log.Info(ctx, "server_starting", "Starting auth service")
	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			log.Error(ctx, "server_error", "Failed to start auth service", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for range ticker.C {
			if !db.IsHealthy(ctx) {
				log.Error(ctx, "db_health_error", "Database connection is not healthy", nil)
				if err := db.Reconnect(ctx); err != nil {
					log.Error(ctx, "db_reconnect_error", "Failed to reconnect to the database", err)
				} else {
					log.Info(ctx, "db_reconnected", "Successfully reconnected to the database")
				}
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info(ctx, "shutdown_initiated", "Received shutdown signal, stopping service")

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Error(ctx, "shutdown_error", "Failed to stop auth service", err)
	}

	log.Info(ctx, "shutdown_complete", "Auth service stopped successfully")
	wg.Wait()
}
