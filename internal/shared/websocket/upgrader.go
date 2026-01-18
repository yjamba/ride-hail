package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader is the default WebSocket upgrader configuration.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// AuthTimeout returns the authentication timeout duration.
func AuthTimeout() int {
	return int(authTimeout.Seconds())
}

// PingPeriod returns the ping period duration in seconds.
func PingPeriod() int {
	return int(pingPeriod.Seconds())
}

// PongWait returns the pong wait duration in seconds.
func PongWait() int {
	return int(pongWait.Seconds())
}
