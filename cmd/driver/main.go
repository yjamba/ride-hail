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
	"ride-hail/internal/shared/broker/messages"
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
	logger.NewLogger("driver-service", getEnv("LOG_LEVEL", "info"))

	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "ride-hail")
	dbPassword := getEnv("POSTGRES_PASSWORD", "ride-hail")
	dbName := getEnv("POSTGRES_DB", "ride-hail-db")
	dbSSL := getEnv("POSTGRES_SSLMODE", "disable")

	dbConfig := postgres.NewDBConfig(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)

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

	// RabbitMQ configuration - используем NewBrokerConfigFromEnv() как в ride service
	brokerConfig := broker.NewBrokerConfigFromEnv()

	slog.Info("connecting to RabbitMQ",
		"host", brokerConfig.Host,
		"port", brokerConfig.Port,
		"vhost", brokerConfig.VHost)

	rabbit := rabbitmq.NewRMQ(brokerConfig)
	if err := rabbit.Connect(ctx); err != nil {
		slog.Error("failed to connect to message broker", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("connected to the message broker successfully")

	// Declare exchanges (как в ride service)
	if err := rabbit.DeclareExchanges(messages.ExchangeDriverTopic, "topic", true, false, false, false, nil); err != nil {
		slog.Error("failed to declare driver exchange", "err", err.Error())
		os.Exit(1)
	}
	if err := rabbit.DeclareExchanges(messages.ExchangeLocationFanout, "fanout", true, false, false, false, nil); err != nil {
		slog.Error("failed to declare location exchange", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("exchanges declared successfully")

	// Declare queues
	queues := []broker.QueueConfig{
		{
			Name:       messages.QueueDriverMatching,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Exchange:   messages.ExchangeDriverTopic,
			RoutingKey: "driver.matching",
		},
		{
			Name:       messages.QueueDriverResponses,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Exchange:   messages.ExchangeDriverTopic,
			RoutingKey: "driver.response.*",
		},
	}

	if err := rabbit.DeclareQueues(queues); err != nil {
		slog.Error("failed to declare queues", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("queues declared successfully")

	app := driver.NewApp(db, rabbit)
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
	slog.Info("received shutdown signal")

	cancel()

	if err := app.Stop(ctx); err != nil {
		slog.Error("failed to stop application", "err", err.Error())
	}
	if err := rabbit.Close(ctx); err != nil {
		slog.Error("failed to close RabbitMQ connection", "err", err.Error())
	}
	if err := db.Close(); err != nil {
		slog.Error("failed to close database connection", "err", err.Error())
	}

	slog.Info("driver service stopped successfully")
	wg.Wait()
}
