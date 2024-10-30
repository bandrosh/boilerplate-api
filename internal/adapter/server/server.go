package server

import (
	"context"
	"fmt"

	"boilerplate-api/config"
)

type Server struct {
	config config.Configuration
}

func NewServer(config config.Configuration) *Server {
	return &Server{config: config}
}

func (s *Server) Run(ctx context.Context, cancel context.CancelFunc) func() error {
	return func() error {
		defer cancel()

		_ = fmt.Sprintf("Starting server %v", ctx)

		return nil
	}
}
