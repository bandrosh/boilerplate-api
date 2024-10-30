package main

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"

	"boilerplate-api/config"
	"boilerplate-api/infrastructure/logger"
	"boilerplate-api/infrastructure/telemetry"
	"boilerplate-api/internal/adapter/server"
)

func main() {
	cfg, err := config.LoadAppConfig("./scripts/config")
	if err != nil {
		logger.Fatal(logger.FatalError, "unable to load configurations")
	}

	var g multierror.Group

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	ctx = context.WithValue(ctx, "service-name", "payment-api")

	tracer, err := telemetry.NewTraceProvider(ctx, cfg)
	if err != nil {
		logger.Fatal(logger.FatalError, "unable to create tracer")
	}

	defer func() {
		if tracer != nil {
			if err := tracer.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Fatal(logger.FatalError, "unable to shutdown tracer")
			}
		}
	}()

	s := server.NewServer(cfg)
	g.Go(s.Run(ctx, stop))

	if err := g.Wait().ErrorOrNil(); err != nil {
		logger.Fatal(logger.ServerError, err)
	}
}
