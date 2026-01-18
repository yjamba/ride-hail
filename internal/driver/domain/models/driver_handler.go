package models

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type OfflineResponse struct {
	Status         string         `json:"status"`
	SessionID      string         `json:"session_id"`
	SessionSummary SessionSummary `json:"session_summary"`
	Message        string         `json:"message"`
}

type StartDriveRequest struct {
	RideID   string   `json:"ride_id"`
	Location Location `json:"driver_location"`
}

type CompleteRideRequest struct {
	RideID                string   `json:"ride_id"`
	Location              Location `json:"final_location"`
	ActualDistanceKm      float64  `json:"actual_distance_km"`
	ActualDurationMinutes int      `json:"actual_duration_minutes"`
}
