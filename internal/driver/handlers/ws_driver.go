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

// DriverHub manages WebSocket connections for drivers.
var DriverHub *websocket.Hub

func init() {
	DriverHub = websocket.NewHub()
	go DriverHub.Run()
}

// UserClaims represents JWT claims for users.
type UserClaims struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RideOffer represents a ride offer sent to drivers.
type RideOffer struct {
	Type                string       `json:"type"`
	OfferID             string       `json:"offer_id"`
	RideID              string       `json:"ride_id"`
	RideNumber          string       `json:"ride_number"`
	PickupLocation      LocationInfo `json:"pickup_location"`
	DestinationLocation LocationInfo `json:"destination_location"`
	EstimatedFare       float64      `json:"estimated_fare"`
	DriverEarnings      float64      `json:"driver_earnings"`
	DistanceToPickupKm  float64      `json:"distance_to_pickup_km"`
	EstimatedDuration   int          `json:"estimated_ride_duration_minutes"`
	ExpiresAt           string       `json:"expires_at"`
}

// LocationInfo represents location with address.
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	Notes     string  `json:"notes,omitempty"`
}

// RideDetails represents ride details sent after driver acceptance.
type RideDetails struct {
	Type           string       `json:"type"`
	RideID         string       `json:"ride_id"`
	PassengerName  string       `json:"passenger_name"`
	PassengerPhone string       `json:"passenger_phone"`
	PickupLocation LocationInfo `json:"pickup_location"`
}

// RideResponse represents a driver's response to a ride offer.
type RideResponse struct {
	Type            string          `json:"type"`
	OfferID         string          `json:"offer_id"`
	RideID          string          `json:"ride_id"`
	Accepted        bool            `json:"accepted"`
	CurrentLocation *LocationCoords `json:"current_location,omitempty"`
}

// LocationCoords represents coordinates.
type LocationCoords struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// LocationUpdate represents a location update from driver.
type LocationUpdate struct {
	Type           string  `json:"type"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	AccuracyMeters float64 `json:"accuracy_meters,omitempty"`
	SpeedKmh       float64 `json:"speed_kmh,omitempty"`
	HeadingDegrees float64 `json:"heading_degrees,omitempty"`
}

// ResponseHandler is a callback for driver ride responses.
type ResponseHandler func(driverID string, response RideResponse)

// LocationHandler is a callback for driver location updates.
type LocationHandler func(driverID string, location LocationUpdate)

// DriverWSHandlerConfig holds configuration for the driver WebSocket handler.
type DriverWSHandlerConfig struct {
	SecretKey        []byte
	OnRideResponse   ResponseHandler
	OnLocationUpdate LocationHandler
}

// DriverWSHandler handles WebSocket connections for drivers.
func DriverWSHandler(config DriverWSHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		driverID := r.PathValue("driver_id")
		if driverID == "" {
			http.Error(w, "driver_id is required", http.StatusBadRequest)
			return
		}

		conn, err := websocket.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := websocket.NewClient(DriverHub, conn, driverID, "DRIVER")

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

		// Handle messages
		messageHandler := createDriverMessageHandler(client, config, authDone)

		DriverHub.Register(client)

		go client.WritePump()
		client.ReadPump(messageHandler)
	}
}

func createDriverMessageHandler(client *websocket.Client, config DriverWSHandlerConfig, authDone chan bool) websocket.MessageHandler {
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
			handleDriverAuth(c, msg, config.SecretKey, authDone)

		case "ping":
			_ = c.SendJSON(map[string]any{"type": "pong"})

		case "ride_response":
			if !c.IsAuthenticated() {
				_ = c.SendJSON(map[string]any{
					"type":    "error",
					"message": "Not authenticated",
				})
				return
			}
			handleRideResponse(c, message, config.OnRideResponse)

		case "location_update":
			if !c.IsAuthenticated() {
				_ = c.SendJSON(map[string]any{
					"type":    "error",
					"message": "Not authenticated",
				})
				return
			}
			handleLocationUpdate(c, message, config.OnLocationUpdate)

		default:
			if !c.IsAuthenticated() {
				_ = c.SendJSON(map[string]any{
					"type":    "error",
					"message": "Not authenticated",
				})
				return
			}
		}
	}
}

func handleDriverAuth(client *websocket.Client, msg map[string]any, secretKey []byte, authDone chan bool) {
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
			"message": "Token does not match driver ID",
		})
		return
	}

	// Verify role
	if claims.Role != "DRIVER" {
		_ = client.SendJSON(map[string]any{
			"type":    "auth_error",
			"message": "Invalid role for driver connection",
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

func handleRideResponse(client *websocket.Client, message []byte, handler ResponseHandler) {
	var response RideResponse
	if err := json.Unmarshal(message, &response); err != nil {
		_ = client.SendJSON(map[string]any{
			"type":    "error",
			"message": "Invalid ride response format",
		})
		return
	}

	if handler != nil {
		handler(client.UserID, response)
	}

	_ = client.SendJSON(map[string]any{
		"type":    "ride_response_received",
		"ride_id": response.RideID,
	})
}

func handleLocationUpdate(client *websocket.Client, message []byte, handler LocationHandler) {
	var update LocationUpdate
	if err := json.Unmarshal(message, &update); err != nil {
		_ = client.SendJSON(map[string]any{
			"type":    "error",
			"message": "Invalid location update format",
		})
		return
	}

	if handler != nil {
		handler(client.UserID, update)
	}
}

// SendRideOfferToDriver sends a ride offer to a driver.
func SendRideOfferToDriver(driverID string, offer RideOffer) error {
	offer.Type = "ride_offer"
	return DriverHub.SendJSONToUser(driverID, offer)
}

// SendRideDetailsToDriver sends ride details to a driver after acceptance.
func SendRideDetailsToDriver(driverID string, details RideDetails) error {
	details.Type = "ride_details"
	return DriverHub.SendJSONToUser(driverID, details)
}

// NotifyDriverRideCancelled notifies a driver that the ride was cancelled.
func NotifyDriverRideCancelled(driverID, rideID, reason string) error {
	return DriverHub.SendJSONToUser(driverID, map[string]any{
		"type":    "ride_cancelled",
		"ride_id": rideID,
		"reason":  reason,
	})
}

// IsDriverConnected checks if a driver is connected via WebSocket.
func IsDriverConnected(driverID string) bool {
	return DriverHub.IsConnected(driverID)
}

// GetConnectedDrivers returns a list of connected driver IDs.
func GetConnectedDrivers() []string {
	return DriverHub.GetConnectedUsers()
}
