package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/bandrosh/boilerplate-api/internal/platform/config"
)

type Server struct {
	httpServer *http.Server
	cfg        config.HTTP
	log        *slog.Logger
}

func New(cfg config.HTTP, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		cfg: cfg,
		log: log,
		httpServer: &http.Server{
			Addr:         net.JoinHostPort("", strconv.Itoa(cfg.Port)),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	s.log.Info("http server listening", slog.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	s.log.Info("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}
