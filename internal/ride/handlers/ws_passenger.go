package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ride-hail/internal/shared/websocket"

	"github.com/golang-jwt/jwt/v5"
)

// PassengerHub manages WebSocket connections for passengers.
var PassengerHub *websocket.Hub

func init() {
	PassengerHub = websocket.NewHub()
	go PassengerHub.Run()
}

// PassengerWSMessage represents messages sent to passengers.
type PassengerWSMessage struct {
	Type      string `json:"type"`
	RideID    string `json:"ride_id,omitempty"`
	RideNum   string `json:"ride_number,omitempty"`
	Status    string `json:"status,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// RideStatusUpdate represents a ride status update message.
type RideStatusUpdate struct {
	Type       string      `json:"type"`
	RideID     string      `json:"ride_id"`
	RideNumber string      `json:"ride_number,omitempty"`
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	DriverInfo *DriverInfo `json:"driver_info,omitempty"`
}

// DriverInfo represents driver information sent to passengers.
type DriverInfo struct {
	DriverID string       `json:"driver_id"`
	Name     string       `json:"name"`
	Rating   float64      `json:"rating"`
	Vehicle  *VehicleInfo `json:"vehicle,omitempty"`
}

// VehicleInfo represents vehicle information.
type VehicleInfo struct {
	Make  string `json:"make"`
	Model string `json:"model"`
	Color string `json:"color"`
	Plate string `json:"plate"`
}

// DriverLocationUpdate represents a driver location update message.
type DriverLocationUpdate struct {
	Type                string   `json:"type"`
	RideID              string   `json:"ride_id"`
	DriverLocation      Location `json:"driver_location"`
	EstimatedArrival    string   `json:"estimated_arrival,omitempty"`
	DistanceToPickupKm  float64  `json:"distance_to_pickup_km,omitempty"`
	DistanceRemainingKm float64  `json:"distance_remaining_km,omitempty"`
}

// Location represents a geographic location.
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// PassengerWSHandler handles WebSocket connections for passengers.
func PassengerWSHandler(secretKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		passengerID := r.PathValue("passenger_id")
		if passengerID == "" {
			http.Error(w, "passenger_id is required", http.StatusBadRequest)
			return
		}

		conn, err := websocket.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := websocket.NewClient(PassengerHub, conn, passengerID, "PASSENGER")

		// Start authentication timeout
		authDone := make(chan bool, 1)
		go func() {
			select {
			case <-authDone:
				return
			case <-time.After(5 * time.Second):
				if !client.IsAuthenticated() {
					_ = conn.WriteJSON(map[string]any{
						"type":    "error",
						"message": "Authentication timeout",
					})
					conn.Close()
				}
			}
		}()

		// Handle authentication in the first message
		messageHandler := createPassengerMessageHandler(client, secretKey, authDone)

		PassengerHub.Register(client)

		go client.WritePump()
		client.ReadPump(messageHandler)
	}
}

func createPassengerMessageHandler(client *websocket.Client, secretKey []byte, authDone chan bool) websocket.MessageHandler {
	return func(c *websocket.Client, message []byte) {
		var msg map[string]any
		if err := json.Unmarshal(message, &msg); err != nil {
			_ = c.SendJSON(map[string]any{
				"type":    "error",
				"message": "Invalid JSON",
			})
			return
		}

		msgType, _ := msg["type"].(string)

		switch msgType {
		case "auth":
			handlePassengerAuth(c, msg, secretKey, authDone)
		case "ping":
			_ = c.SendJSON(map[string]any{"type": "pong"})
		default:
			if !c.IsAuthenticated() {
				_ = c.SendJSON(map[string]any{
					"type":    "error",
					"message": "Not authenticated",
				})
				return
			}
			// Handle other message types if needed
		}
	}
}

func handlePassengerAuth(client *websocket.Client, msg map[string]any, secretKey []byte, authDone chan bool) {
	tokenStr, _ := msg["token"].(string)
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		_ = client.SendJSON(map[string]any{
			"type":    "auth_error",
			"message": "Invalid token",
		})
		return
	}

	// Verify user ID matches
	if claims.UserId != client.UserID {
		_ = client.SendJSON(map[string]any{
			"type":    "auth_error",
			"message": "Token does not match passenger ID",
		})
		return
	}

	// Verify role
	if claims.Role != "PASSENGER" {
		_ = client.SendJSON(map[string]any{
			"type":    "auth_error",
			"message": "Invalid role for passenger connection",
		})
		return
	}

	client.SetAuthenticated(true)
	select {
	case authDone <- true:
	default:
	}

	_ = client.SendJSON(map[string]any{
		"type":    "auth_success",
		"message": "Successfully authenticated",
	})
}

// UserClaims represents JWT claims for users.
type UserClaims struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// SendRideStatusToPassenger sends a ride status update to a passenger.
func SendRideStatusToPassenger(passengerID string, update RideStatusUpdate) error {
	update.Type = "ride_status_update"
	return PassengerHub.SendJSONToUser(passengerID, update)
}

// SendDriverLocationToPassenger sends driver location to a passenger.
func SendDriverLocationToPassenger(passengerID string, update DriverLocationUpdate) error {
	update.Type = "driver_location_update"
	return PassengerHub.SendJSONToUser(passengerID, update)
}

// NotifyPassengerRideMatched notifies a passenger that their ride was matched.
func NotifyPassengerRideMatched(passengerID, rideID, rideNumber string, driver *DriverInfo) error {
	return SendRideStatusToPassenger(passengerID, RideStatusUpdate{
		Type:       "ride_status_update",
		RideID:     rideID,
		RideNumber: rideNumber,
		Status:     "MATCHED",
		DriverInfo: driver,
	})
}

// NotifyPassengerDriverArrived notifies a passenger that their driver arrived.
func NotifyPassengerDriverArrived(passengerID, rideID string) error {
	return SendRideStatusToPassenger(passengerID, RideStatusUpdate{
		Type:    "ride_status_update",
		RideID:  rideID,
		Status:  "ARRIVED",
		Message: "Your driver has arrived at the pickup location",
	})
}

// NotifyPassengerRideStarted notifies a passenger that their ride started.
func NotifyPassengerRideStarted(passengerID, rideID string) error {
	return SendRideStatusToPassenger(passengerID, RideStatusUpdate{
		Type:   "ride_status_update",
		RideID: rideID,
		Status: "IN_PROGRESS",
	})
}

// NotifyPassengerRideCompleted notifies a passenger that their ride completed.
func NotifyPassengerRideCompleted(passengerID, rideID string, finalFare float64) error {
	return PassengerHub.SendJSONToUser(passengerID, map[string]any{
		"type":       "ride_status_update",
		"ride_id":    rideID,
		"status":     "COMPLETED",
		"final_fare": finalFare,
		"message":    "Your ride has been completed. Thank you!",
	})
}

// NotifyPassengerRideCancelled notifies a passenger that their ride was cancelled.
func NotifyPassengerRideCancelled(passengerID, rideID, reason string) error {
	return SendRideStatusToPassenger(passengerID, RideStatusUpdate{
		Type:    "ride_status_update",
		RideID:  rideID,
		Status:  "CANCELLED",
		Message: reason,
	})
}
