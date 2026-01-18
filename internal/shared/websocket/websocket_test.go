package websocket

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.clients == nil {
		t.Fatal("expected non-nil clients map")
	}
}

func TestHub_Count(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	if hub.Count() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.Count())
	}
}

func TestHub_IsConnected(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	if hub.IsConnected("user123") {
		t.Error("expected user to not be connected")
	}
}

func TestHub_GetConnectedUsers(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	users := hub.GetConnectedUsers()
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestHub_SendToUser_NotFound(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	err := hub.SendToUser("nonexistent", []byte("test"))
	if err != ErrClientNotFound {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
}

func TestAuthTimeout(t *testing.T) {
	timeout := AuthTimeout()
	if timeout != 5 {
		t.Errorf("expected auth timeout 5, got %d", timeout)
	}
}

func TestPingPeriod(t *testing.T) {
	period := PingPeriod()
	if period != 30 {
		t.Errorf("expected ping period 30, got %d", period)
	}
}

func TestPongWait(t *testing.T) {
	wait := PongWait()
	if wait != 60 {
		t.Errorf("expected pong wait 60, got %d", wait)
	}
}

func TestClientConstants(t *testing.T) {
	if writeWait != 10*time.Second {
		t.Errorf("expected writeWait 10s, got %v", writeWait)
	}
	if pongWait != 60*time.Second {
		t.Errorf("expected pongWait 60s, got %v", pongWait)
	}
	if pingPeriod != 30*time.Second {
		t.Errorf("expected pingPeriod 30s, got %v", pingPeriod)
	}
	if maxMessageSize != 512 {
		t.Errorf("expected maxMessageSize 512, got %d", maxMessageSize)
	}
	if authTimeout != 5*time.Second {
		t.Errorf("expected authTimeout 5s, got %v", authTimeout)
	}
}
