package models

import "time"

type Overview struct {
	Time               time.Time
	Metrics            *Metrics
	DriverDistribution *DriverDistribution
	Hotspots           []Hotspots
}

type Metrics struct {
	ActiveRides                int
	AvailableDrivers           int
	BusyDrivers                int
	TotalRidesToday            int
	TotalRevenueToday          float64
	AverageWaitTimeMinutes     float32
	AverageRideDurationMinutes float32
	CancellationRate           float32
}

type DriverDistribution struct {
	Economy int
	Premium int
	XL      int
}

type Hotspots struct {
	Location     string
	ActiveRides  int
	WaitingRides int
}
