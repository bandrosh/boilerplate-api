package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App           App
	HTTP          HTTP
	Log           Log
	AWS           AWS
	Observability Observability
}

type App struct {
	Name string `env:"APP_NAME" envDefault:"boilerplate-api"`
	Env  string `env:"APP_ENV"  envDefault:"local"`
}

func (a App) IsProduction() bool { return a.Env == "production" }

type HTTP struct {
	Port            int           `env:"HTTP_PORT"           envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT"   envDefault:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT"  envDefault:"10s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT"   envDefault:"120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT"    envDefault:"15s"`
}

type Log struct {
	Level  string `env:"LOG_LEVEL"  envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

type AWS struct {
	Region          string `env:"AWS_REGION"            envDefault:"us-east-1"`
	Endpoint        string `env:"AWS_ENDPOINT_URL"      envDefault:"http://127.0.0.1:4566"`
	AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"     envDefault:"test"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY" envDefault:"test"`
	DynamoTable     string `env:"DYNAMODB_TABLE"        envDefault:"boilerplate"`
}

type Observability struct {
	Enabled      bool   `env:"OTEL_ENABLED"                 envDefault:"true"`
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"127.0.0.1:4317"`
	ServiceName  string `env:"OTEL_SERVICE_NAME"            envDefault:"boilerplate-api"`
}

func Load() (Config, error) {

	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
