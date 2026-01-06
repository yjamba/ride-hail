package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
)

const (
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[36m"
	reset  = "\033[0m"
)

func InitLogger(level string) *slog.Logger {
	levelSlog := slog.LevelDebug
	switch level {
	case "debug":
		levelSlog = slog.LevelDebug
	case "warn":
		levelSlog = slog.LevelWarn
	case "error":
		levelSlog = slog.LevelError
	case "info":
		levelSlog = slog.LevelInfo
	default:
		fmt.Println("i don't know this log level, set to info")
		levelSlog = slog.LevelInfo
	}
	opts := PrettyHandleOptions{
		SlogOpt: slog.HandlerOptions{
			Level: levelSlog,
		},
	}
	var handler slog.Handler = NewPrettyHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))

	return slog.Default()
}

type PrettyHandleOptions struct {
	SlogOpt slog.HandlerOptions
}

type PrettyHandler struct {
	slog.Handler
	l *log.Logger
}

func (p *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"
	switch r.Level {
	case slog.LevelDebug:
		level = blue + level + reset
	case slog.LevelInfo:
		level = green + level + reset
	case slog.LevelWarn:
		level = yellow + level + reset
	case slog.LevelError:
		level = red + level + reset
	}
	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = fmt.Sprint(a.Value.Any())
		return true
	})

	b, err := json.MarshalIndent(fields, " ", " ")
	if err != nil {
		return err
	}
	timeStr := r.Time.Format("[2006-01-02 15:04:05]")

	if string(b) == "{}" {
		p.l.Println(timeStr, level, r.Message)
	} else {
		p.l.Println(timeStr, level, r.Message, string(b))
	}
	return nil
}

func NewPrettyHandler(out io.Writer, opts PrettyHandleOptions) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewJSONHandler(out, &opts.SlogOpt),
		l:       log.New(out, "", 0),
	}
}
