// Package config loads and validates application configuration from the
// environment. A local .env file is loaded as a convenience when running from
// the IDE; in real environments variables are injected by the platform.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config aggregates every configuration section the application needs.
type Config struct {
	App           App
	HTTP          HTTP
	Log           Log
	AWS           AWS
	Observability Observability
}

// App holds high-level application metadata.
type App struct {
	Name string `env:"APP_NAME" envDefault:"boilerplate-api"`
	Env  string `env:"APP_ENV"  envDefault:"local"`
}

// IsProduction reports whether the app runs in a production-like environment.
func (a App) IsProduction() bool { return a.Env == "production" }

// HTTP holds the inbound HTTP server settings.
type HTTP struct {
	Port            int           `env:"HTTP_PORT"           envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"   envDefault:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"  envDefault:"10s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT"   envDefault:"120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT"    envDefault:"15s"`
}

// Log holds logging settings.
type Log struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// AWS holds AWS SDK / DynamoDB settings. Endpoint points to LocalStack in
// local development and is left empty in real AWS environments.
type AWS struct {
	Region          string `env:"AWS_REGION"            envDefault:"us-east-1"`
	Endpoint        string `env:"AWS_ENDPOINT_URL"      envDefault:"http://localhost:4566"`
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"     envDefault:"test"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY" envDefault:"test"`
	DynamoTable     string `env:"DYNAMODB_TABLE"        envDefault:"boilerplate"`
}

// Observability holds OpenTelemetry settings.
type Observability struct {
	Enabled      bool   `env:"OTEL_ENABLED"                 envDefault:"true"`
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4317"`
	ServiceName  string `env:"OTEL_SERVICE_NAME"            envDefault:"boilerplate-api"`
}

// Load reads configuration from the environment. It first tries to load a
// local .env file (ignored if absent) and then parses env vars into Config.
func Load() (Config, error) {
	// Best-effort: a missing .env is not an error (e.g. in production).
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
