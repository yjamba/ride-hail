package ws

import "ride-hail/internal/driver/domain/ports"

type WSNotifier struct {
	hub *Hub
}

func NewWSNotifier(h *Hub) *WSNotifier {
	return &WSNotifier{hub: h}
}

// Notify implements ports.Notifier by broadcasting JSON via the Hub.
func (w *WSNotifier) Notify(event interface{}) error {
	if w.hub == nil {
		return nil
	}
	return w.hub.BroadcastJSON(event)
}

// Ensure WSNotifier implements ports.Notifier
var _ ports.Notifier = (*WSNotifier)(nil)
