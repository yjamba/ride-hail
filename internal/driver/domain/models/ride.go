package models

import "time"

type Ride struct {
	ID            string
	RideNumber    string
	PassengerID   string
	DriverID      string
	VehicleType   string
	Status        RideStatus
	EstimatedFare float64
	FinalFare     float64
	CreatedAt     time.Time
}
