package service

import (
	"context"
	"encoding/json"
	"errors"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/domain/ports"
	"ride-hail/internal/shared/broker/messages"
	"ride-hail/internal/shared/logger"
	"time"
)

type RideService struct {
	repo      ports.RideRepository
	publisher ports.Publish
	logger    *logger.Logger
	secretKey []byte
}

func NewRideService(repo ports.RideRepository, publisher ports.Publish, log *logger.Logger, secretKey []byte) *RideService {
	return &RideService{
		repo:      repo,
		publisher: publisher,
		logger:    log,
		secretKey: secretKey,
	}
}

func (s *RideService) CreateRide(ctx context.Context, cmd models.CreateRideCommand) (*models.Ride, error) {
	if err := validateLanLon(cmd.Pickup.Latitude, cmd.Pickup.Longitude); err != nil {
		s.logError(ctx, "validation_error", "Invalid pickup coordinates", err)
		return nil, err
	}
	if err := validateLanLon(cmd.Destination.Latitude, cmd.Destination.Longitude); err != nil {
		s.logError(ctx, "validation_error", "Invalid destination coordinates", err)
		return nil, err
	}

	ride := &models.Ride{
		PassengerID:         cmd.PassengerID,
		VehicleType:         cmd.VehicleType,
		Status:              models.RideStatusRequested,
		PickupLocation:      cmd.Pickup,
		DestinationLocation: cmd.Destination,
		RequestedAt:         time.Now(),
	}
	if err := s.repo.CreateRide(ctx, ride); err != nil {
		s.logError(ctx, "db_error", "Failed to create ride in database", err)
		return nil, err
	}

	// Add ride_id to context for logging
	ctx = logger.WithRideID(ctx, ride.ID)

	s.logInfo(ctx, "ride_requested", "New ride request created", map[string]interface{}{
		"passenger_id": ride.PassengerID,
		"vehicle_type": ride.VehicleType,
		"pickup":       ride.PickupLocation,
		"destination":  ride.DestinationLocation,
	})

	// Publish ride match request to message broker
	if err := s.publishRideMatchRequest(ctx, ride); err != nil {
		s.logError(ctx, "publish_error", "Failed to publish ride match request", err)
		// Don't fail the ride creation if publishing fails
	}

	return ride, nil
}

func (s *RideService) GetRideById(ctx context.Context, rideID, passengerID string) (models.Ride, error) {
	ride, err := s.repo.GetRide(ctx, rideID)
	if err != nil {
		return models.Ride{}, err
	}
	return ride, nil
}

func (s *RideService) GetRideByStatus(ctx context.Context, passengerID string, status string) ([]models.Ride, error) {
	rides, err := s.repo.ListByStatus(ctx, passengerID, status)
	if err != nil {
		return nil, err
	}
	return rides, nil
}

func (s *RideService) UpdateRideStatus(ctx context.Context, rideID string, status string) error {
	validStatuses := []string{"REQUESTED", "IN_PROGRESS", "COMPLETED", "CANCELLED"}
	if !contains(validStatuses, status) {
		return errors.New("invalid ride status")
	}
	if err := s.repo.UpdateStatus(ctx, rideID, status); err != nil {
		s.logError(ctx, "db_error", "Failed to update ride status", err)
		return err
	}

	ctx = logger.WithRideID(ctx, rideID)
	s.logInfo(ctx, "status_changed", "Ride status updated", map[string]interface{}{
		"new_status": status,
	})
	return nil
}

func (s *RideService) CloseRide(ctx context.Context, id string, reason string) error {
	if err := s.repo.CloseRide(ctx, id, reason); err != nil {
		return err
	}
	return nil
}

func validateLanLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}

	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}

	return nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// publishRideMatchRequest publishes a ride match request to the message broker
func (s *RideService) publishRideMatchRequest(ctx context.Context, ride *models.Ride) error {
	if s.publisher == nil {
		return nil
	}

	msg := messages.RideMatchRequest{
		RideID:     ride.ID,
		RideNumber: ride.RideNumber,
		PickupLocation: messages.Coordinate{
			Lat:     ride.PickupLocation.Latitude,
			Lng:     ride.PickupLocation.Longitude,
			Address: ride.PickupLocation.Address,
		},
		Destination: messages.Coordinate{
			Lat:     ride.DestinationLocation.Latitude,
			Lng:     ride.DestinationLocation.Longitude,
			Address: ride.DestinationLocation.Address,
		},
		RideType:       string(ride.VehicleType),
		EstimatedFare:  getEstimatedFare(ride.EstimatedFare),
		MaxDistanceKm:  10.0, // Default max distance for driver matching
		TimeoutSeconds: 60,   // Default timeout for driver response
		RequestedAt:    ride.RequestedAt,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := messages.RideRequestRoutingKey(string(ride.VehicleType))
	return s.publisher.Publish(ctx, messages.ExchangeRideTopic, routingKey, body)
}

// publishRideStatusUpdate publishes a ride status update to the message broker
func (s *RideService) publishRideStatusUpdate(ctx context.Context, ride *models.Ride) error {
	if s.publisher == nil {
		return nil
	}

	msg := messages.RideStatusUpdate{
		RideID:    ride.ID,
		DriverID:  ride.DriverID,
		Status:    string(ride.Status),
		Timestamp: time.Now(),
	}

	if ride.Status == models.RideStatusCompleted && ride.FinalFare != nil {
		msg.FinalFare = ride.FinalFare
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := messages.RideStatusRoutingKey(string(ride.Status))
	return s.publisher.Publish(ctx, messages.ExchangeRideTopic, routingKey, body)
}

func getEstimatedFare(fare *float64) float64 {
	if fare == nil {
		return 0
	}
	return *fare
}

// Logging helper methods
func (s *RideService) logInfo(ctx context.Context, action, message string, extra interface{}) {
	if s.logger != nil {
		s.logger.InfoWithFields(ctx, action, message, extra)
	}
}

func (s *RideService) logError(ctx context.Context, action, message string, err error) {
	if s.logger != nil {
		s.logger.Error(ctx, action, message, err)
	}
}

func (s *RideService) logDebug(ctx context.Context, action, message string) {
	if s.logger != nil {
		s.logger.Debug(ctx, action, message)
	}
}
