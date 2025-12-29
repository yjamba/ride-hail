package messages

import (
	"fmt"
	"time"
)

// Exchanges
const (
	ExchangeRideTopic      = "ride_topic"
	ExchangeDriverTopic    = "driver_topic"
	ExchangeLocationFanout = "location_fanout"
)

// Queues (recommended names from README)
const (
	// ride_topic
	QueueRideRequests = "ride_requests"
	QueueRideStatus   = "ride_status"

	// driver_topic
	QueueDriverMatching  = "driver_matching"
	QueueDriverResponses = "driver_responses"
	QueueDriverStatus    = "driver_status"

	// location_fanout
	QueueLocationUpdatesRide = "location_updates"
)

// Routing key helpers
func RideRequestRoutingKey(rideType string) string  { return fmt.Sprintf("ride.request.%s", rideType) }
func RideStatusRoutingKey(status string) string     { return fmt.Sprintf("ride.status.%s", status) }
func DriverResponseRoutingKey(rideID string) string { return fmt.Sprintf("driver.response.%s", rideID) }
func DriverStatusRoutingKey(driverID string) string { return fmt.Sprintf("driver.status.%s", driverID) }

// ---------- Common / Nested types ----------

type Coordinate struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address,omitempty"`
}

type VehicleInfo struct {
	Make  string `json:"make,omitempty"`
	Model string `json:"model,omitempty"`
	Color string `json:"color,omitempty"`
	Plate string `json:"plate,omitempty"`
}

type DriverInfo struct {
	DriverID string       `json:"driver_id,omitempty"`
	Name     string       `json:"name,omitempty"`
	Rating   float64      `json:"rating,omitempty"`
	Vehicle  *VehicleInfo `json:"vehicle,omitempty"`
}

// ---------- Ride service -> Driver matching (outgoing) ----------

// RideMatchRequest is published to ride_topic with routing key ride.request.{ride_type}
type RideMatchRequest struct {
	RideID         string     `json:"ride_id"`
	RideNumber     string     `json:"ride_number,omitempty"`
	PickupLocation Coordinate `json:"pickup_location"`
	Destination    Coordinate `json:"destination_location"`
	RideType       string     `json:"ride_type"`
	EstimatedFare  float64    `json:"estimated_fare,omitempty"`
	MaxDistanceKm  float64    `json:"max_distance_km,omitempty"`
	TimeoutSeconds int        `json:"timeout_seconds,omitempty"`
	CorrelationID  string     `json:"correlation_id,omitempty"`
	RequestedAt    time.Time  `json:"requested_at,omitempty"`
}

// ---------- Driver -> Ride service (incoming) ----------

// DriverMatchResponse is published by driver service to driver_topic with routing key driver.response.{ride_id}
type DriverMatchResponse struct {
	RideID                  string      `json:"ride_id"`
	DriverID                string      `json:"driver_id"`
	Accepted                bool        `json:"accepted"`
	EstimatedArrivalMinutes int         `json:"estimated_arrival_minutes,omitempty"`
	DriverLocation          *Coordinate `json:"driver_location,omitempty"`
	DriverInfo              *DriverInfo `json:"driver_info,omitempty"`
	EstimatedArrival        *time.Time  `json:"estimated_arrival,omitempty"`
	CorrelationID           string      `json:"correlation_id,omitempty"`
}

// ---------- Location updates (fanout) ----------

// LocationUpdate is broadcast on location_fanout
type LocationUpdate struct {
	DriverID  string     `json:"driver_id"`
	RideID    string     `json:"ride_id,omitempty"`
	Location  Coordinate `json:"location"`
	SpeedKmh  float64    `json:"speed_kmh,omitempty"`
	Heading   float64    `json:"heading_degrees,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// ---------- Ride status updates (ride_topic) ----------

type RideStatusUpdate struct {
	RideID        string    `json:"ride_id"`
	DriverID      string    `json:"driver_id"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp,omitempty"`
	FinalFare     *float64  `json:"final_fare,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Message       string    `json:"message,omitempty"`
}

// ---------- Driver status updates (driver_topic) ----------

type DriverStatusUpdate struct {
	DriverID  string    `json:"driver_id"`
	Status    string    `json:"status"`
	RideID    string    `json:"ride_id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}
