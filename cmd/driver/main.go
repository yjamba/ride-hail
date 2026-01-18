package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ride-hail/internal/driver"
	"ride-hail/internal/shared/broker"
	"ride-hail/internal/shared/broker/rabbitmq"
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
	log := logger.NewLogger("auth-service", getEnv("LOG_LEVEL", "info"))
	_ = log

	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "ride-hail")
	dbPassword := getEnv("POSTGRES_PASSWORD", "ride-hail")
	dbName := getEnv("POSTGRES_DB", "ride-hail-db")
	dbSSL := getEnv("POSTGRES_SSLMODE", "disable")

	dbConfig := postgres.NewDBConfig(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)
	rabbitHost := getEnv("RABBITMQ_HOST", "rabbitmq") // "rabbitmq" вместо "localhost"
	rabbitPort := getEnv("RABBITMQ_PORT", "5672")
	rabbitUser := getEnv("RABBITMQ_USER", "guest")
	rabbitPassword := getEnv("RABBITMQ_PASSWORD", "guest_password") // "guest" вместо "guest_password"
	rabbitVHost := getEnv("RABBITMQ_VHOST", "/")

	brokerConfig := broker.NewBrokerConfig(rabbitHost, rabbitPort, rabbitUser, rabbitPassword, rabbitVHost)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	db := postgres.NewDB(dbConfig)

	if err := db.Connect(ctx); err != nil {
		slog.Error("failed to connect to db", "err", err.Error())
		os.Exit(1)
	}

	slog.Info("connected to the database successfully")

	rabbit := rabbitmq.NewRMQ(brokerConfig)
	if err := rabbit.Connect(ctx); err != nil {
		slog.Error("failed to connect to message broker", "err", err.Error())
		os.Exit(1)
	}

	slog.Info("connected to the message broker successfully")

	app := driver.NewApp()
	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			slog.Error("failed to start server", "err", err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !db.IsHealthy(ctx) {
					slog.Error("database connection is not healthy")
					if err := db.Reconnect(ctx); err != nil {
						slog.Error("failed to reconnect to the database", "error", err.Error())
					} else {
						slog.Info("successfully reconnected to the database")
					}
				}

				if !rabbit.IsHealthy(ctx) {
					slog.Error("message broker connection is not healthy")
					if err := rabbit.Reconnect(ctx); err != nil {
						slog.Error("failed to reconnect to the message broker", "error", err.Error())
					} else {
						slog.Info("successfully reconnected to the message broker")
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	cancel()

	if err := app.Stop(ctx); err != nil {
		slog.Error("failed to stop application", "err", err.Error())
		return
	}

	wg.Wait()
}
