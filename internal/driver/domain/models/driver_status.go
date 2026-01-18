package models

type DriverStatus string

const (
	Offline   DriverStatus = "OFFLINE"
	Available DriverStatus = "AVAILABLE"
	Busy      DriverStatus = "BUSY"
	EnRoute   DriverStatus = "EN_ROUTE" // в пути
)

func (s DriverStatus) String() string {
	return string(s)
}

func (s DriverStatus) IsValid() bool {
	switch s {
	case Offline, Available, Busy, EnRoute:
		return true
	default:
		return false
	}
}

func ParseDriverStatus(s string) (DriverStatus, bool) {
	ds := DriverStatus(s)
	return ds, ds.IsValid()
}

type RideStatus string

const (
	RideStatusRequested  RideStatus = "REQUESTED"
	RideStatusMatched    RideStatus = "MATCHED"
	RideStatusEnRoute    RideStatus = "EN_ROUTE"
	RideStatusArrived    RideStatus = "ARRIVED"
	RideStatusInProgress RideStatus = "IN_PROGRESS"
	RideStatusCompleted  RideStatus = "COMPLETED"
	RideStatusCancelled  RideStatus = "CANCELLED"
)

func (s RideStatus) String() string {
	return string(s)
}
