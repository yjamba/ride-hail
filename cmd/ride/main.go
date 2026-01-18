package main

import (
	"context"
	"os"
	"os/signal"
	"ride-hail/internal/ride"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/ride/repository"
	"ride-hail/internal/shared/broker"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/broker/rabbitmq"
	"ride-hail/internal/shared/config"
	"ride-hail/internal/shared/logger"
	"strconv"
	"sync"
	"syscall"
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
	log := logger.NewLogger("ride-service", getEnv("LOG_LEVEL", "info"))
	ctx := context.Background()

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
	if err := rmq.Connect(ctx); err != nil {
		log.Error(ctx, "rabbitmq_connect_error", "Failed to connect to RabbitMQ", err)
		os.Exit(1)
	}
	log.Info(ctx, "rabbitmq_connected", "Successfully connected to RabbitMQ")

	if err := rmq.DeclareExchanges(messages.ExchangeRideTopic, "topic", true, false, false, false, nil); err != nil {
		log.Error(ctx, "exchange_declare_error", "Failed to declare ride exchange", err)
		os.Exit(1)
	}
	if err := rmq.DeclareExchanges(messages.ExchangeDriverTopic, "topic", true, false, false, false, nil); err != nil {
		log.Error(ctx, "exchange_declare_error", "Failed to declare driver exchange", err)
		os.Exit(1)
	}
	if err := rmq.DeclareExchanges(messages.ExchangeLocationFanout, "fanout", true, false, false, false, nil); err != nil {
		log.Error(ctx, "exchange_declare_error", "Failed to declare location exchange", err)
		os.Exit(1)
	}
	log.Info(ctx, "exchanges_declared", "All RabbitMQ exchanges declared successfully")

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
		log.Error(ctx, "queue_declare_error", "Failed to declare queues", err)
		os.Exit(1)
	}
	log.Info(ctx, "queues_declared", "All RabbitMQ queues declared successfully")

	db := &repository.DB{}

	app := ride.NewApp(config, db, rmq, log)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			log.Error(ctx, "server_error", "Failed to start ride service", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info(ctx, "shutdown_initiated", "Received shutdown signal, stopping service")

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Error(ctx, "shutdown_error", "Failed to stop ride service", err)
	}
	if err := rmq.Close(shutdownCtx); err != nil {
		log.Error(ctx, "rabbitmq_close_error", "Failed to close RabbitMQ connection", err)
	}

	log.Info(ctx, "shutdown_complete", "Ride service stopped successfully")
	wg.Wait()
}
