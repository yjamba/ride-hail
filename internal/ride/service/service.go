package service

import (
	"context"
	"errors"
	"ride-hail/internal/ride/domain/models"
	"ride-hail/internal/ride/repository"
	"time"
)

type RideService struct {
	repo repository.RideRepo
	// publisher
	// logger
}

func NewRideService(repo repository.RideRepo) *RideService {
	return &RideService{repo: repo}
}

func (s *RideService) CreateRide(ctx context.Context, cmd models.CreateRideCommand) (*models.RideDB, error) {
	// if err := validateLanLon(message.PickupLat, message.PickupLon); err != nil { // проверка начальной точки
	// 	return nil, err
	// }
	// if err := validateLanLon(message.DestLat, message.DestLon); err != nil { // проверка конечной точки
	// 	return nil, err
	// }

	ride := &models.RideDB{
		PassengerID:           cmd.PassengerID,
		Status:                "REQUESTED",
		PickupCoordinate:      cmd.CordinatePickup,
		DestinationCoordinate: cmd.CordinateDest,
		RequestedAt: time.Now(),
	}
	if err := s.repo.CreateRide(ctx, ride); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *RideService) GetRideById(ctx context.Context, id string) error {
	return nil
}

func (s *RideService) GetRideByStatus(ctx context.Context, passengerID string) error {
	return nil
}

func (s *RideService) UpdateRideStatus(ctx context.Context, ride *models.CreateRideCommand) error {
	return nil
}

func (s *RideService) CloseRide(ctx context.Context, id string) error {
	return nil
	//удалять нельзя просто закрыть поездку
}

func validateLanLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("invalid lat")
	}

	if lon < -180 || lon > 180 {
		return errors.New("invalid lon")
	}

	return nil
}

// ('IN_PROGRESS'), -- Ride is currently in progress
// ('COMPLETED'),   -- Ride has been successfully completed
// ('CANCELLED')    -- Ride was cancelled
