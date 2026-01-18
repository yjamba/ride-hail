package ws

import (
	"log/slog"
	"net/http"
)

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(h *Hub) *WSHandler {
	return &WSHandler{hub: h}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	driverID := r.PathValue("driver_id")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Info("upgrade error: %v", err)
		return
	}
	c := &connection{
		ws:       ws,
		send:     make(chan []byte, 256),
		hub:      h.hub,
		driverID: driverID,
	}
	// register connection with hub so it will receive broadcasts
	if h.hub != nil {
		h.hub.Register(c)
	}
	go c.writePump()
	go c.readPump()
}
