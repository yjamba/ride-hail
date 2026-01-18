package service

import (
	"context"
	"errors"
	"log"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/domain/ports"
)

type RideService struct {
	repo      ports.RideRepository
	secretKey []byte
	// publisher
	// logger
}

func NewRideService(repo ports.RideRepository, secretKey []byte) *RideService {
	return &RideService{
		repo:      repo,
		secretKey: secretKey,
	}
}

func (s *RideService) CreateRide(ctx context.Context, cmd models.CreateRideCommand) (*models.Ride, error) {
	if err := validateLanLon(cmd.Pickup.Latitude, cmd.Pickup.Longitude); err != nil { // проверка начальной точки
		return nil, err
	}
	if err := validateLanLon(cmd.Destination.Latitude, cmd.Destination.Longitude); err != nil { // проверка конечной точки
		return nil, err
	}

	ride := &models.Ride{
		PassengerID:         cmd.PassengerID,
		Status:              models.RideStatusRequested,
		PickupLocation:      cmd.Pickup,
		DestinationLocation: cmd.Destination,
	}
	if err := s.repo.CreateRide(ctx, ride); err != nil {
		return nil, err
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
		return err
	}
	log.Printf("Ride %s status updated to %s", rideID, status)
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

// ('IN_PROGRESS'), -- Ride is currently in progress
// ('COMPLETED'),   -- Ride has been successfully completed
// ('CANCELLED')    -- Ride was cancelled
