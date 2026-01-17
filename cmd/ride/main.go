package ride

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"ride-hail/internal/ride"
	"ride-hail/internal/ride/handlers"
	"ride-hail/internal/ride/repository"
	"sync"
	"syscall"
)

func main() {
	config := &handlers.ServerConfig{
		Addr: "localhost",
		Port: 4001,
	}
	db := &repository.DB{}

	app := ride.NewApp(config, db)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Start(context.Background()); err != nil {
			slog.Error("failed to start auth service", "error", err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := app.Stop(ctx); err != nil {
		slog.Error("failed to stop ride service", "error", err.Error())
	}
	wg.Wait()
}
