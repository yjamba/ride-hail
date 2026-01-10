package handlers

import (
	"context"
	"errors"
	"net"
	"net/http"
)

type Server struct {
	server *http.Server
	config ServerConfig

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(config ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.server = &http.Server{
		Addr:    s.config.GetAddr(),
		Handler: nil,
		BaseContext: func(l net.Listener) context.Context {
			return s.ctx
		},
	}

	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	return errors.New("Server is nil")
}
