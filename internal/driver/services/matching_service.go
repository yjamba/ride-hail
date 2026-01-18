package services

import (
	"context"
	"encoding/json"
	"log"

	"ride-hail/internal/driver/domain/ports"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/broker/rabbitmq"
)

type MatchingService struct {
	notifier   ports.Notifier
	consume    ports.Consume
	driverRepo ports.DriverRepository
}

func NewMatchingService(notifier ports.Notifier, consume ports.Consume, driverRepo ports.DriverRepository) *MatchingService {
	return &MatchingService{
		notifier:   notifier,
		consume:    consume,
		driverRepo: driverRepo,
	}
}

func (m *MatchingService) Start(ctx context.Context) error {
	// Start consuming from the matching queue.
	// The Consume interface returns a channel of raw message bytes.
	ch, err := m.consume.Consume(messages.ExchangeDriverTopic, "ride.request")
	if err != nil {
		return err
	}

	// Process messages in a goroutine.
	go m.processMessages(ctx, ch)
	return nil
}

// processMessages reads from the message channel and handles each message.
func (m *MatchingService) processMessages(ctx context.Context, ch <-chan rabbitmq.Message) {
	for msg := range ch {
		if err := m.handleMessage(ctx, msg); err != nil {
			log.Printf("error handling message: %v", err)
		}
	}
}

// handleMessage processes a single message from the queue.
// Decodes JSON and routes to a driver via notifier or broadcasts.
// handleMessage processes a RideMatchRequest from the queue.
func (m *MatchingService) handleMessage(ctx context.Context, msg rabbitmq.Message) error {
	var req messages.RideMatchRequest
	if err := json.Unmarshal(msg.Body(), &req); err != nil {
		log.Printf("failed to unmarshal RideMatchRequest: %v", err)
		return err
	}

	// Find available drivers nearby the pickup location
	drivers, err := m.driverRepo.FindAvailableDriversNearby(
		ctx,
		req.PickupLocation.Lat,
		req.PickupLocation.Lng,
		req.RideType,
		5000,
	)
	if err != nil {
		log.Printf("failed to find available drivers: %v", err)
		return err
	}

	if len(drivers) == 0 {
		log.Printf("no available drivers for ride %s", req.RideID)
		return nil
	}

	// Select the first (nearest) available driver
	matchedDriver := drivers[0]

	// Notify the matched driver
	evt := map[string]interface{}{
		"type":           "ride_match",
		"ride_id":        req.RideID,
		"driver_id":      matchedDriver.ID,
		"pickup":         req.PickupLocation,
		"destination":    req.Destination,
		"ride_type":      req.RideType,
		"estimated_fare": req.EstimatedFare,
		"correlation_id": req.CorrelationID,
	}

	return m.notifier.Notify(evt)
}
