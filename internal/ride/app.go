package ride

import (
"context"

"ride-hail/internal/ride/domain/ports"
"ride-hail/internal/ride/handlers"
"ride-hail/internal/ride/repository"
"ride-hail/internal/ride/service"
"ride-hail/internal/shared/broker/rabbitmq"
"ride-hail/internal/shared/logger"
"ride-hail/internal/shared/postgres"
)

type App struct {
config    *handlers.ServerConfig
db        *postgres.Database
rmq       *rabbitmq.RMQ
publisher ports.Publish
logger    *logger.Logger

server *handlers.Server
}

func NewApp(config *handlers.ServerConfig, db *postgres.Database, rmq *rabbitmq.RMQ, log *logger.Logger) *App {
return &App{
config:    config,
db:        db,
rmq:       rmq,
publisher: rmq,
logger:    log,
}
}

func (a *App) Start(ctx context.Context) error {
repo := repository.NewRideRepo(a.db)

svc := service.NewRideService(repo, a.publisher, a.logger, []byte("supersecretkey"))
handler := handlers.NewRideHandler(svc)

a.server = handlers.NewServer(handler, a.config)

if a.logger != nil {
a.logger.Info(ctx, "server_starting", "Ride service starting")
}

if err := a.server.Start(ctx); err != nil {
return err
}

return nil
}

func (a *App) Stop(ctx context.Context) error {
if a.server != nil {
return a.server.Stop(ctx)
}

return nil
}
