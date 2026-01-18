package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"ride-hail/internal/ride"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/shared/broker"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/broker/rabbitmq"
	"ride-hail/internal/shared/config"
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

	port := 4001
	if p, err := strconv.Atoi(getEnv("RIDE_SERVICE_PORT", "4001")); err == nil {
		port = p
	}

	config := &handlers.ServerConfig{
		Addr: getEnv("RIDE_SERVICE_HOST", "0.0.0.0"),
		Port: port,
	}

	brokerConfig := broker.NewBrokerConfigFromEnv()
	rmq := rabbitmq.NewRMQ(brokerConfig)
	if err := rmq.Connect(context.Background()); err != nil {
		slog.Error("failed to connect to rabbitmq", "error", err.Error())
		os.Exit(1)
	}

	if err := rmq.DeclareExchanges(messages.ExchangeRideTopic, "topic", true, false, false, false, nil); err != nil {
		slog.Error("failed to declare ride exchange", "error", err.Error())
		os.Exit(1)
	}
	if err := rmq.DeclareExchanges(messages.ExchangeDriverTopic, "topic", true, false, false, false, nil); err != nil {
		slog.Error("failed to declare driver exchange", "error", err.Error())
		os.Exit(1)
	}
	if err := rmq.DeclareExchanges(messages.ExchangeLocationFanout, "fanout", true, false, false, false, nil); err != nil {
		slog.Error("failed to declare location exchange", "error", err.Error())
		os.Exit(1)
	}

	queues := []broker.QueueConfig{
		{
			Name:       messages.QueueRideRequests,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Exchange:   messages.ExchangeRideTopic,
			RoutingKey: "ride.request.*",
		},
		{
			Name:       messages.QueueRideStatus,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Exchange:   messages.ExchangeRideTopic,
			RoutingKey: "ride.status.*",
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
		{
			Name:       messages.QueueDriverStatus,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Exchange:   messages.ExchangeDriverTopic,
			RoutingKey: "driver.status.*",
		},
		{
			Name:       messages.QueueLocationUpdatesRide,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Exchange:   messages.ExchangeLocationFanout,
			RoutingKey: "",
		},
	}

	if err := rmq.DeclareQueues(queues); err != nil {
		slog.Error("failed to declare queues", "error", err.Error())
		os.Exit(1)
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

	app := ride.NewApp(config, db, rmq)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Start(context.Background()); err != nil {
			slog.Error("failed to start auth service", "error", err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if !db.IsHealthy(ctx) {
			slog.Error("postgres is not healthy")
			if err := db.Reconnect(ctx); err != nil {
				slog.Error("failed to reconnect to postgres", "error", err.Error())
			} else {
				slog.Info("reconnected to postgres")
			}
		}

		if !rmq.IsHealthy(ctx) {
			slog.Error("rabbitmq is not healthy")
			if err := rmq.Reconnect(ctx); err != nil {
				slog.Error("failed to reconnect to rabbitmq", "error", err.Error())
			} else {
				slog.Info("reconnected to rabbitmq")
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	if err := app.Stop(ctx); err != nil {
		slog.Error("failed to stop ride service", "error", err.Error())
	}
	if err := rmq.Close(ctx); err != nil {
		slog.Error("failed to close rabbitmq", "error", err.Error())
	}
	wg.Wait()
}
