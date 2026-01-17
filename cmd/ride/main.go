package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"ride-hail/internal/ride"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/ride/repository"
	"ride-hail/internal/shared/broker"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/broker/rabbitmq"
	"ride-hail/internal/shared/config"
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
	db := &repository.DB{}

	app := ride.NewApp(config, db)

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
		slog.Error("failed to stop ride service", "error", err.Error())
	}
	if err := rmq.Close(ctx); err != nil {
		slog.Error("failed to close rabbitmq", "error", err.Error())
	}
	wg.Wait()
}
