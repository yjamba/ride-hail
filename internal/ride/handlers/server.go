package handlers

import (
	"context"
	"log/slog"
	"net"
	"net/http"
)

type Server struct {
	rideHandler  *RideHandler
	serverConfig *ServerConfig
	secretKey    []byte

	server *http.Server

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(rideHandler *RideHandler, serverConfig *ServerConfig, secretKey []byte) *Server {
	return &Server{
		rideHandler:  rideHandler,
		serverConfig: serverConfig,
		secretKey:    secretKey,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.server = &http.Server{
		Addr:    s.serverConfig.GetAddr(),
		Handler: RegisterRoutes(s.rideHandler, s.secretKey),
		BaseContext: func(l net.Listener) context.Context {
			return s.ctx
		},
	}

	slog.Info("starting ride server", "addr", s.serverConfig.GetAddr())
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("stopping ride server")
	if s.cancel != nil {
		s.cancel()
	}

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}
