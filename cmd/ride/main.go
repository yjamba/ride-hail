package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"ride-hail/internal/ride"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/shared/broker"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/broker/rabbitmq"
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

	// Initialize structured logger
	log := logger.NewLogger("ride-service", getEnv("LOG_LEVEL", "info"))

	port := 3004
	if p, err := strconv.Atoi(getEnv("RIDE_SERVICE_PORT", "3004")); err == nil {
		port = p
	}

	serverConfig := &handlers.ServerConfig{
		Addr: getEnv("RIDE_SERVICE_HOST", "0.0.0.0"),
		Port: port,
	}

	// Database configuration
	dbHost := getEnv("POSTGRES_HOST", "postgres")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "ride-hail")
	dbPassword := getEnv("POSTGRES_PASSWORD", "ride-hail")
	dbName := getEnv("POSTGRES_DB", "ride-hail-db")
	dbSSL := getEnv("POSTGRES_SSLMODE", "disable")

	dbConfig := postgres.NewDBConfig(dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)
	log.InfoWithFields(ctx, "db_connecting", "Connecting to database", map[string]interface{}{
		"host": dbHost,
		"port": dbPort,
		"db":   dbName,
	})

	db := postgres.NewDB(dbConfig)
	if err := db.Connect(ctx); err != nil {
		log.Error(ctx, "db_connect_error", "Failed to connect to database", err)
		os.Exit(1)
	}
	log.Info(ctx, "db_connected", "Successfully connected to database")

	// RabbitMQ configuration
	brokerConfig := broker.NewBrokerConfigFromEnv()
	rmq := rabbitmq.NewRMQ(brokerConfig)
	if err := rmq.Connect(ctx); err != nil {
		log.Error(ctx, "rabbitmq_connect_error", "Failed to connect to RabbitMQ", err)
		os.Exit(1)
	}
	log.Info(ctx, "rabbitmq_connected", "Successfully connected to RabbitMQ")

	// Declare exchanges
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

	// Declare queues
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

	// Secret key for JWT
	secretKey := []byte(getEnv("JWT_SECRET", "supersecretkey"))

	app := ride.NewApp(serverConfig, db, rmq, log, secretKey)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			log.Error(ctx, "server_error", "Failed to start ride service", err)
		}
	}()

	// Health check goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Second)
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

				if !rmq.IsHealthy(ctx) {
					log.Error(ctx, "rabbitmq_health_error", "RabbitMQ connection is not healthy", nil)
					if err := rmq.Reconnect(ctx); err != nil {
						log.Error(ctx, "rabbitmq_reconnect_error", "Failed to reconnect to RabbitMQ", err)
					} else {
						log.Info(ctx, "rabbitmq_reconnected", "Successfully reconnected to RabbitMQ")
					}
				}
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info(ctx, "shutdown_initiated", "Received shutdown signal, stopping service")

	cancel()

	if err := app.Stop(ctx); err != nil {
		log.Error(ctx, "shutdown_error", "Failed to stop ride service", err)
	}
	if err := rmq.Close(ctx); err != nil {
		log.Error(ctx, "rabbitmq_close_error", "Failed to close RabbitMQ connection", err)
	}
	if err := db.Close(); err != nil {
		log.Error(ctx, "db_close_error", "Failed to close database connection", err)
	}

	log.Info(ctx, "shutdown_complete", "Ride service stopped successfully")
	wg.Wait()
}
