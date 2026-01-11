package models

type DriverStatus string

const (
	Offline   DriverStatus = "OFFLINE"
	Available DriverStatus = "AVAILABLE"
	Busy      DriverStatus = "BUSY"
	EnRoute   DriverStatus = "EN_ROUTE"
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
