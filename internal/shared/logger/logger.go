package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Log levels
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelError = "ERROR"
)

// Context keys for request tracing
type ctxKey string

const (
	CtxRequestID ctxKey = "request_id"
	CtxRideID    ctxKey = "ride_id"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Service   string      `json:"service"`
	Action    string      `json:"action"`
	Message   string      `json:"message"`
	Hostname  string      `json:"hostname"`
	RequestID string      `json:"request_id,omitempty"`
	RideID    string      `json:"ride_id,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Extra     interface{} `json:"extra,omitempty"`
}

// ErrorInfo contains error details for ERROR level logs
type ErrorInfo struct {
	Msg   string `json:"msg"`
	Stack string `json:"stack,omitempty"`
}

// Logger is a structured JSON logger
type Logger struct {
	service  string
	hostname string
	output   io.Writer
	level    string
	mu       sync.Mutex
}

// NewLogger creates a new structured logger
func NewLogger(service string, level string) *Logger {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	validLevel := LevelInfo
	switch strings.ToUpper(level) {
	case LevelDebug:
		validLevel = LevelDebug
	case LevelInfo:
		validLevel = LevelInfo
	case LevelError:
		validLevel = LevelError
	}

	return &Logger{
		service:  service,
		hostname: hostname,
		output:   os.Stdout,
		level:    validLevel,
	}
}

// SetOutput sets the output writer for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// shouldLog checks if the message should be logged based on level
func (l *Logger) shouldLog(level string) bool {
	levelOrder := map[string]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelError: 2,
	}
	return levelOrder[level] >= levelOrder[l.level]
}

// log writes a log entry
func (l *Logger) log(ctx context.Context, level, action, message string, err error, extra interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Service:   l.service,
		Action:    action,
		Message:   message,
		Hostname:  l.hostname,
	}

	// Extract request_id and ride_id from context
	if ctx != nil {
		if requestID, ok := ctx.Value(CtxRequestID).(string); ok {
			entry.RequestID = requestID
		}
		if rideID, ok := ctx.Value(CtxRideID).(string); ok {
			entry.RideID = rideID
		}
	}

	// Add error info for ERROR level
	if err != nil && level == LevelError {
		entry.Error = &ErrorInfo{
			Msg:   err.Error(),
			Stack: getStackTrace(),
		}
	}

	// Add extra fields if provided
	if extra != nil {
		entry.Extra = extra
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		fmt.Fprintf(l.output, `{"timestamp":"%s","level":"ERROR","service":"%s","action":"log_error","message":"failed to marshal log entry"}`+"\n",
			time.Now().UTC().Format(time.RFC3339), l.service)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, action, message string) {
	l.log(ctx, LevelDebug, action, message, nil, nil)
}

// DebugWithFields logs a debug message with extra fields
func (l *Logger) DebugWithFields(ctx context.Context, action, message string, extra interface{}) {
	l.log(ctx, LevelDebug, action, message, nil, extra)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, action, message string) {
	l.log(ctx, LevelInfo, action, message, nil, nil)
}

// InfoWithFields logs an info message with extra fields
func (l *Logger) InfoWithFields(ctx context.Context, action, message string, extra interface{}) {
	l.log(ctx, LevelInfo, action, message, nil, extra)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, action, message string, err error) {
	l.log(ctx, LevelError, action, message, err, nil)
}

// ErrorWithFields logs an error message with extra fields
func (l *Logger) ErrorWithFields(ctx context.Context, action, message string, err error, extra interface{}) {
	l.log(ctx, LevelError, action, message, err, extra)
}

// WithRequestID returns a new context with request_id
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, CtxRequestID, requestID)
}

// WithRideID returns a new context with ride_id
func WithRideID(ctx context.Context, rideID string) context.Context {
	return context.WithValue(ctx, CtxRideID, rideID)
}

// getStackTrace returns a simplified stack trace
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(4, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()
		// Skip runtime and standard library frames
		if !strings.Contains(frame.File, "runtime/") {
			fmt.Fprintf(&builder, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
		}
		if !more {
			break
		}
	}
	return builder.String()
}
