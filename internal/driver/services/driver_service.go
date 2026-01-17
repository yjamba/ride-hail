package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"ride-hail/internal/driver/domain/models"
	"ride-hail/internal/driver/domain/ports"
)

type DriverService struct {
	repo ports.DriverRepository
}

func NewDriverService(repo ports.DriverRepository) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) GoOffline(ctx context.Context, driverID string) error {
	// Validate driverID
	if driverID == "" {
		return errors.New("driverID cannot be empty")
	}

	// Get the current driver status
	stat, err := s.repo.GetById(ctx, driverID)
	if err != nil {
		log.Printf("Failed to get driver by ID %s: %v", driverID, err)
		return fmt.Errorf("failed to get driver by ID %s: %w", driverID, err)
	}

	// Check if the driver is already offline
	if stat.Status != models.Available {
		return fmt.Errorf(
			"cannot go offline: current status %s",
			stat.Status,
		)
	}

	// Update the driver's status to OFFLINE
	if err := s.repo.UpdateStatus(ctx, driverID, models.Offline); err != nil {
		log.Printf("Failed to set driver %s offline: %v", driverID, err)
		return fmt.Errorf("failed to set driver %s offline: %w", driverID, err)
	}

	log.Printf("Driver %s is now offline", driverID)
	return nil
}

func (s *DriverService) GoOnline(ctx context.Context, driverID string) error {
	// Validate driverID
	if driverID == "" {
		return errors.New("driverID cannot be empty")
	}

	// Get the current driver status
	stat, err := s.repo.GetById(ctx, driverID)
	if err != nil {
		log.Printf("Failed to get driver by ID %s: %v", driverID, err)
		return fmt.Errorf("failed to get driver by ID %s: %w", driverID, err)
	}

	// Check if the driver is already online
	if stat.Status != models.Offline {
		return fmt.Errorf(
			"cannot go online: current status %s",
			stat.Status,
		)
	}

	// Update the driver's status to AVAILABLE
	if err := s.repo.UpdateStatus(ctx, driverID, models.Available); err != nil {
		log.Printf("Failed to set driver %s online: %v", driverID, err)
		return fmt.Errorf("failed to set driver %s online: %w", driverID, err)
	}

	log.Printf("Driver %s is now online", driverID)
	return nil
}

func (s *DriverService) UpdateLocation(ctx context.Context, driverID string, lat, lon float64) error {
	// Validate driverID
	if driverID == "" {
		return errors.New("driverID cannot be empty")
	}

	// Validate latitude and longitude
	if err := validateLatLon(lat, lon); err != nil {
		log.Printf("Invalid location for driver %s: %v", driverID, err)
		return err
	}

	stat, err := s.repo.GetById(ctx, driverID)
	if err != nil {
		return err
	}

	if stat.Status == models.Offline {
		return errors.New("cannot update location: driver offline")
	}

	// Update the driver's location in the database
	if err := s.repo.UpdateDriverLocation(ctx, driverID, lat, lon); err != nil {
		log.Printf("Failed to update location for driver %s: %v", driverID, err)
		return fmt.Errorf("failed to update location for driver %s: %w", driverID, err)
	}

	log.Printf("Driver %s location updated to lat: %f, lon: %f", driverID, lat, lon)
	return nil
}

func (s *DriverService) AcceptRide(ctx context.Context, driverID, rideID string) error {
	// Validate rideID
	if rideID == "" {
		return errors.New("rideID cannot be empty")
	}

	// Update the ride's status to MATCHED
	if err := s.repo.UpdateRideStatus(ctx, rideID, models.RideStatusMatched); err != nil {
		log.Printf("Driver %s failed to accept ride %s: %v", driverID, rideID, err)
		return fmt.Errorf("failed to accept ride %s by driver %s: %w", rideID, driverID, err)
	}

	log.Printf("Driver %s accepted ride %s", driverID, rideID)
	return nil
}

func (s *DriverService) CompleteRide(ctx context.Context, driverID, rideID string) error {
	// Validate rideID
	if rideID == "" {
		return errors.New("rideID cannot be empty")
	}

	// Update the ride's status to COMPLETED
	if err := s.repo.UpdateRideStatus(ctx, rideID, models.RideStatusCompleted); err != nil {
		log.Printf("Driver %s failed to complete ride %s: %v", driverID, rideID, err)
		return fmt.Errorf("failed to complete ride %s by driver %s: %w", rideID, driverID, err)
	}

	log.Printf("Driver %s completed ride %s", driverID, rideID)
	return nil
}

func (s *DriverService) StartRide(ctx context.Context, driverID, rideID string) error {
	// Validate rideID
	if rideID == "" {
		return errors.New("rideID cannot be empty")
	}

	// Update the ride's status to IN_PROGRESS
	if err := s.repo.UpdateRideStatus(ctx, rideID, models.RideStatusInProgress); err != nil {
		log.Printf("Driver %s failed to start ride %s: %v", driverID, rideID, err)
		return fmt.Errorf("failed to start ride %s by driver %s: %w", rideID, driverID, err)
	}

	log.Printf("Driver %s started ride %s", driverID, rideID)
	return nil
}

func validateLatLon(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}
