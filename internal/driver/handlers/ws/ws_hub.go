package ws

import (
	"encoding/json"
	"log"
)

type driverMessage struct {
	driverID string
	msg      []byte
}

// ...existing code...
type Hub struct {
	clients      map[*connection]bool
	register     chan *connection
	unregister   chan *connection
	broadcast    chan []byte
	sendToDriver chan driverMessage
	done         chan struct{}
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*connection]bool),
		register:     make(chan *connection),
		unregister:   make(chan *connection),
		broadcast:    make(chan []byte, 256),
		sendToDriver: make(chan driverMessage, 256),
		done:         make(chan struct{}),
	}
}

// Run processes hub events. Call Start() to run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			log.Printf("ws: registered connection (total=%d)", len(h.clients))

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				// close send to signal writer goroutine to exit
				close(c.send)
				log.Printf("ws: unregistered connection (total=%d)", len(h.clients))
			}

		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
					// delivered
				default:
					// client is not reading; drop client
					close(c.send)
					delete(h.clients, c)
					log.Printf("ws: dropped slow client (total=%d)", len(h.clients))
				}
			}

		case dm := <-h.sendToDriver:
			for c := range h.clients {
				if c.driverID == dm.driverID {
					select {
					case c.send <- dm.msg:
						// sent
					default:
						// slow client -> drop
						close(c.send)
						delete(h.clients, c)
					}
				}
			}
		case <-h.done:
			// cleanup
			for c := range h.clients {
				close(c.send)
				delete(h.clients, c)
			}
			return
		}
	}
}

// Start runs the hub loop in a new goroutine.
func (h *Hub) Start() {
	go h.Run()
}

// Stop signals the hub to stop and cleans up clients.
func (h *Hub) Stop() {
	close(h.done)
}

// Register adds a connection to the hub.
func (h *Hub) Register(c *connection) {
	h.register <- c
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(c *connection) {
	h.unregister <- c
}

// Broadcast sends a message to all connected clients (non-blocking).
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
		// drop if broadcast channel is full
	}
}

func (h *Hub) BroadcastRaw(b []byte) {
	h.Broadcast(b) // uses existing Broadcast implementation
}

func (h *Hub) SendToDriver(driverID string, b []byte) {
	select {
	case h.sendToDriver <- driverMessage{driverID: driverID, msg: b}:
	default:
		// drop if channel full
	}
}

// SendToDriverJSON marshals v and sends to the specified driver.
func (h *Hub) SendToDriverJSON(driverID string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	h.SendToDriver(driverID, b)
	return nil
}

// BroadcastJSON marshals v to JSON and broadcasts it. Returns error if marshal fails.
func (h *Hub) BroadcastJSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	h.Broadcast(b)
	return nil
}
