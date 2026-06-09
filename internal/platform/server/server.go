// Package server owns the HTTP server lifecycle, including graceful shutdown.
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

// Server wraps *http.Server with start/stop helpers.
type Server struct {
	httpServer *http.Server
	cfg        config.HTTP
	log        *slog.Logger
}

// New builds the HTTP server from config and a handler.
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

// Start begins serving and blocks until the server stops. ErrServerClosed is
// returned as nil so a graceful shutdown is not treated as a failure.
func (s *Server) Start() error {
	s.log.Info("http server listening", slog.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

// Shutdown gracefully drains in-flight requests, bounded by SHUTDOWN_TIMEOUT.
func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	s.log.Info("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}
