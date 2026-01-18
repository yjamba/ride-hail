package main

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = config.LoadEnv()
	getEnv := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	log := logger.NewLogger("admin-service", getEnv("LOG_LEVEL", "info"))

	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "ride-hail")
	dbPassword := getEnv("POSTGRES_PASSWORD", "ride-hail")
	dbName := getEnv("POSTGRES_DB", "ride-hail-db")
	dbSSL := getEnv("POSTGRES_SSLMODE", "disable")

	dbConfig := postgres.NewDBConfig(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)
	db := postgres.NewDB(dbConfig)
	if err := db.Connect(ctx); err != nil {
		log.Error(ctx, "db_connect_error", "Failed to connect to postgres", err)
		os.Exit(1)
	}
	log.Info(ctx, "db_connected", "Successfully connected to database")

	port := 3003
	if p, err := strconv.Atoi(getEnv("ADMIN_SERVICE_PORT", "3003")); err == nil {
		port = p
	}

	serverConfig := &handlers.ServerConfig{
		Addr: getEnv("ADMIN_SERVICE_HOST", "0.0.0.0"),
		Port: port,
	}

	app := admin.NewApp(db, serverConfig, log)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !db.IsHealthy(ctx) {
					log.Error(ctx, "db_health_error", "Database connection is not healthy", nil)
					if err := db.Reconnect(ctx); err != nil {
						log.Error(ctx, "db_reconnect_error", "Failed to reconnect to database", err)
					} else {
						log.Info(ctx, "db_reconnected", "Successfully reconnected to database")
					}
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			log.Error(ctx, "server_error", "Failed to start admin service", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info(ctx, "shutdown_initiated", "Received shutdown signal, stopping service")

	cancel()

	if err := app.Stop(ctx); err != nil {
		log.Error(ctx, "shutdown_error", "Failed to stop admin service", err)
	}

	log.Info(ctx, "shutdown_complete", "Admin service stopped successfully")
	wg.Wait()
}
