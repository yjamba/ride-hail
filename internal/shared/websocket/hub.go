package websocket

import (
	"errors"
	"sync"
)

// MessageHandler is a function that handles incoming WebSocket messages.
type MessageHandler func(client *Client, message []byte)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// Registered clients by user ID.
	clients map[string]*Client

	// Inbound messages from clients.
	broadcast chan []byte

	// Register requests from clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Mutex for thread-safe operations.
	mu sync.RWMutex
}

// Common errors.
var (
	ErrClientNotFound   = errors.New("client not found")
	ErrClientBufferFull = errors.New("client send buffer full")
)

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Close existing connection if any
			if existing, ok := h.clients[client.UserID]; ok {
				close(existing.send)
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register registers a client with the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToUser sends a message to a specific user.
func (h *Hub) SendToUser(userID string, message []byte) error {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return ErrClientNotFound
	}

	select {
	case client.send <- message:
		return nil
	default:
		return ErrClientBufferFull
	}
}

// SendJSONToUser sends a JSON message to a specific user.
func (h *Hub) SendJSONToUser(userID string, v any) error {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return ErrClientNotFound
	}

	return client.SendJSON(v)
}

// GetClient returns a client by user ID.
func (h *Hub) GetClient(userID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[userID]
	return client, ok
}

// GetConnectedUsers returns a list of connected user IDs.
func (h *Hub) GetConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// Count returns the number of connected clients.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsConnected checks if a user is connected.
func (h *Hub) IsConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
