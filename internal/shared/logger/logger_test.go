package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNewLogger_CreatesLogger(t *testing.T) {
	l := NewLogger("test-service", "info")
	if l == nil {
		t.Fatal("expected logger, got nil")
	}
	if l.service != "test-service" {
		t.Errorf("expected service 'test-service', got %q", l.service)
	}
}

func TestNewLogger_DefaultsToInfoLevel(t *testing.T) {
	l := NewLogger("test-service", "unknown")
	if l.level != LevelInfo {
		t.Errorf("expected level INFO, got %q", l.level)
	}
}

func TestLogger_Info_WritesJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("ride-service", "debug")
	l.SetOutput(&buf)

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithRideID(ctx, "ride-456")

	l.Info(ctx, "ride_requested", "New ride request received")

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v, output: %s", err, buf.String())
	}

	if entry.Level != LevelInfo {
		t.Errorf("expected level INFO, got %q", entry.Level)
	}
	if entry.Service != "ride-service" {
		t.Errorf("expected service 'ride-service', got %q", entry.Service)
	}
	if entry.Action != "ride_requested" {
		t.Errorf("expected action 'ride_requested', got %q", entry.Action)
	}
	if entry.Message != "New ride request received" {
		t.Errorf("expected message 'New ride request received', got %q", entry.Message)
	}
	if entry.RequestID != "req-123" {
		t.Errorf("expected request_id 'req-123', got %q", entry.RequestID)
	}
	if entry.RideID != "ride-456" {
		t.Errorf("expected ride_id 'ride-456', got %q", entry.RideID)
	}
	if entry.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestLogger_Error_IncludesErrorInfo(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("ride-service", "debug")
	l.SetOutput(&buf)

	ctx := context.Background()
	testErr := errors.New("database connection failed")

	l.Error(ctx, "db_error", "Failed to connect to database", testErr)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if entry.Level != LevelError {
		t.Errorf("expected level ERROR, got %q", entry.Level)
	}
	if entry.Error == nil {
		t.Fatal("expected error info, got nil")
	}
	if entry.Error.Msg != "database connection failed" {
		t.Errorf("expected error msg 'database connection failed', got %q", entry.Error.Msg)
	}
	if entry.Error.Stack == "" {
		t.Error("expected stack trace to be set")
	}
}

func TestLogger_Debug_RespectLogLevel(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("test-service", "info") // INFO level, DEBUG should be skipped
	l.SetOutput(&buf)

	l.Debug(context.Background(), "debug_action", "This should not appear")

	if buf.Len() > 0 {
		t.Errorf("expected no output for DEBUG when level is INFO, got: %s", buf.String())
	}
}

func TestLogger_InfoWithFields_IncludesExtra(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("ride-service", "debug")
	l.SetOutput(&buf)

	extra := map[string]interface{}{
		"pickup_lat":  43.238949,
		"pickup_lng":  76.889709,
		"fare_amount": 1500.0,
	}

	l.InfoWithFields(context.Background(), "fare_calculated", "Fare calculated successfully", extra)

	output := buf.String()
	if !strings.Contains(output, "pickup_lat") {
		t.Errorf("expected output to contain extra fields, got: %s", output)
	}
	if !strings.Contains(output, "fare_amount") {
		t.Errorf("expected output to contain fare_amount, got: %s", output)
	}
}

func TestWithRequestID_AddsToContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-req-id")

	val, ok := ctx.Value(CtxRequestID).(string)
	if !ok || val != "test-req-id" {
		t.Errorf("expected request_id 'test-req-id', got %q", val)
	}
}

func TestWithRideID_AddsToContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithRideID(ctx, "test-ride-id")

	val, ok := ctx.Value(CtxRideID).(string)
	if !ok || val != "test-ride-id" {
		t.Errorf("expected ride_id 'test-ride-id', got %q", val)
	}
}