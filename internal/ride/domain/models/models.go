package models

import "time"

type RideDB struct { // DB
	ID                 string
	RideNumber         string
	PassengerID        string
	DriverID           *string
	Status             string
	VehicleType        *string
	Priority           int
	RequestedAt        time.Time
	MatchedAt          *time.Time
	ArrivedAt          *time.Time
	StartedAt          *time.Time
	CompletedAt        *time.Time
	CancelledAt        *time.Time
	CancellationReason *string
	EstimatedFare      *float64
	FinalFare          *float64

	PickupCoordinateID string
	PickupCoordinate   *Cordinate

	DestinationCoordinate Cordinate
}

type CreateRideCommand struct { // service
	PassengerID string

	CordinatePickup Cordinate

	CordinateDest Cordinate

	RideType string
}

type Cordinate struct {
	Lat     float64
	Lon     float64
	Address string
}
