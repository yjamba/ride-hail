package broker

import "testing"

func TestBrokerConfig_GetConnectionString(t *testing.T) {
	cfg := &BrokerConfig{
		Host:     "localhost",
		Port:     "5672",
		User:     "guest",
		Password: "guest",
	}

	got := cfg.GetConnectionString()
	want := "amqp://guest:guest@localhost:5672/%2f"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
