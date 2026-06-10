package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bandrosh/boilerplate-api/internal/adapter/dynamodb"
	apphttp "github.com/bandrosh/boilerplate-api/internal/adapter/http"
	"github.com/bandrosh/boilerplate-api/internal/adapter/http/handler"
	appuser "github.com/bandrosh/boilerplate-api/internal/application/user"
	"github.com/bandrosh/boilerplate-api/internal/platform/config"
	"github.com/bandrosh/boilerplate-api/internal/platform/logger"
	"github.com/bandrosh/boilerplate-api/internal/platform/observability"
	"github.com/bandrosh/boilerplate-api/internal/platform/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log)
	slog.SetDefault(log)
	log.Info("starting service",
		slog.String("app", cfg.App.Name),
		slog.String("env", cfg.App.Env),
	)

	shutdownOtel, err := observability.Setup(ctx, cfg.Observability, cfg.App.Env)
	if err != nil {
		return err
	}
	defer func() { _ = shutdownOtel(context.Background()) }()

	dynamoClient, err := dynamodb.NewClient(ctx, cfg.AWS)
	if err != nil {
		return err
	}
	log.Info("dynamodb client ready", slog.String("table", cfg.AWS.DynamoTable))

	userRepo := dynamodb.NewUserRepository(dynamoClient, cfg.AWS.DynamoTable)
	userSvc := appuser.NewService(userRepo, log)
	userHandler := handler.NewUserHandler(userSvc)
	healthHandler := handler.NewHealthHandler(dynamodb.NewHealthChecker(dynamoClient, cfg.AWS.DynamoTable))

	router := apphttp.Router{
		Log:    log,
		User:   userHandler,
		Health: healthHandler,
	}.Build()

	srv := server.New(cfg.HTTP, router, log)

	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.Start() }()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info("interrupt received, shutting down gracefully")
		return srv.Shutdown(context.Background())
	}
}
