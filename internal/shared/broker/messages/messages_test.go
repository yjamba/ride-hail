package messages

import "testing"

func TestRideRequestRoutingKey(t *testing.T) {
	tests := []struct {
		rideType string
		expected string
	}{
		{"ECONOMY", "ride.request.ECONOMY"},
		{"PREMIUM", "ride.request.PREMIUM"},
		{"XL", "ride.request.XL"},
	}

	for _, tc := range tests {
		got := RideRequestRoutingKey(tc.rideType)
		if got != tc.expected {
			t.Errorf("RideRequestRoutingKey(%q) = %q, want %q", tc.rideType, got, tc.expected)
		}
	}
}

func TestRideStatusRoutingKey(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"REQUESTED", "ride.status.REQUESTED"},
		{"MATCHED", "ride.status.MATCHED"},
		{"COMPLETED", "ride.status.COMPLETED"},
		{"CANCELLED", "ride.status.CANCELLED"},
	}

	for _, tc := range tests {
		got := RideStatusRoutingKey(tc.status)
		if got != tc.expected {
			t.Errorf("RideStatusRoutingKey(%q) = %q, want %q", tc.status, got, tc.expected)
		}
	}
}

func TestDriverResponseRoutingKey(t *testing.T) {
	rideID := "550e8400-e29b-41d4-a716-446655440000"
	expected := "driver.response.550e8400-e29b-41d4-a716-446655440000"

	got := DriverResponseRoutingKey(rideID)
	if got != expected {
		t.Errorf("DriverResponseRoutingKey(%q) = %q, want %q", rideID, got, expected)
	}
}

func TestDriverStatusRoutingKey(t *testing.T) {
	driverID := "660e8400-e29b-41d4-a716-446655440001"
	expected := "driver.status.660e8400-e29b-41d4-a716-446655440001"

	got := DriverStatusRoutingKey(driverID)
	if got != expected {
		t.Errorf("DriverStatusRoutingKey(%q) = %q, want %q", driverID, got, expected)
	}
}

func TestExchangeConstants(t *testing.T) {
	if ExchangeRideTopic != "ride_topic" {
		t.Errorf("ExchangeRideTopic = %q, want ride_topic", ExchangeRideTopic)
	}
	if ExchangeDriverTopic != "driver_topic" {
		t.Errorf("ExchangeDriverTopic = %q, want driver_topic", ExchangeDriverTopic)
	}
	if ExchangeLocationFanout != "location_fanout" {
		t.Errorf("ExchangeLocationFanout = %q, want location_fanout", ExchangeLocationFanout)
	}
}

func TestQueueConstants(t *testing.T) {
	queues := map[string]string{
		"QueueRideRequests":       QueueRideRequests,
		"QueueRideStatus":         QueueRideStatus,
		"QueueDriverMatching":     QueueDriverMatching,
		"QueueDriverResponses":    QueueDriverResponses,
		"QueueDriverStatus":       QueueDriverStatus,
		"QueueLocationUpdatesRide": QueueLocationUpdatesRide,
	}

	for name, value := range queues {
		if value == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}
