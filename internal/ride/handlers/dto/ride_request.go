package dto

type CreateRideRequest struct {
	PassengerID string  `json:"passenger_id"`
	PickupLat   float64 `json:"pickup_lat"`
	PickupLon   float64 `json:"pickup_lon"`
	PickupAddr  string  `json:"pickup_address"`
	DestLat     float64 `json:"dest_lat"`
	DestLon     float64 `json:"dest_lon"`
	DestAddress string  `json:"dest_address"`
}

type CancelRideRequest struct {
	Reason string `json:"reason"`
}
