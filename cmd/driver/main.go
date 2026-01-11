package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ride-hail/internal/driver"
	"ride-hail/internal/shared/logger"
)

func main() {
	logger.InitLogger("debug")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	app := driver.NewApp()
	go func() {
		defer wg.Done()
		if err := app.Start(ctx); err != nil {
			slog.Error("Failed to start server", "err", err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				slog.Info("Reconnecting to db")
			case <-ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	cancel()

	if err := app.Stop(ctx); err != nil {
		slog.Error("Failed to stop application", "err", err.Error())
		return
	}

	wg.Wait()
}
