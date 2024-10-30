package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Configuration struct {
	Server    Server    `mapstructure:"server"`
	DynamoDB  DynamoDB  `mapstructure:"dynamodb"`
	Telemetry Telemetry `mapstructure:"telemetry"`
}

type Telemetry struct {
	Hostname string `mapstructure:"hostname"`
}

type DynamoDB struct {
	Endpoint string `mapstructure:"endpoint"`
	Region   string `mapstructure:"region"`
}

type Server struct {
	Port int `mapstructure:"port"`
}

func LoadAppConfig(path string) (Configuration, error) {
	if path == "" {
		return Configuration{}, fmt.Errorf("config path is empty")
	}

	viper.AddConfigPath(path)
	viper.SetConfigName("local")
	viper.SetConfigType("yml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError

		if errors.As(err, &configFileNotFoundError) {
			return Configuration{}, fmt.Errorf("config file not found: %s", path)
		}

		return Configuration{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var configuration Configuration

	if err := viper.Unmarshal(&configuration); err != nil {
		return Configuration{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return configuration, nil
}
