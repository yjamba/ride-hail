package main

import (
	"context"
	"log/slog"

	"ride-hail/internal/driver"
	"ride-hail/internal/shared/logger"
)

func main() {
	logger.InitLogger("debug")

	app := driver.NewApp()

	if err := app.Start(context.Background()); err != nil {
		slog.Error("Error")
	}
}
