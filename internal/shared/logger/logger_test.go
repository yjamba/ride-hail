package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestInitLogger_UnknownLevelDefaultsToInfo(t *testing.T) {
	l := InitLogger("unknown")
	if l == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestPrettyHandler_Handle_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	h := NewPrettyHandler(&buf, PrettyHandleOptions{SlogOpt: slog.HandlerOptions{Level: slog.LevelDebug}})

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	rec.AddAttrs(slog.String("request_id", "req-1"))

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected output to contain message, got %q", out)
	}
	if !strings.Contains(out, "request_id") {
		t.Fatalf("expected output to contain attrs, got %q", out)
	}

}