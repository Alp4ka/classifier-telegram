package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	// Core service.
	CoreGRPCAddress string `env:"CORE_GRPC_ADDRESS" envDefault:"localhost:229"`

	// Security.
	APIKey string `env:"API_KEY" envDefault:"7448821770:AAExE-5kWp4ywKle1eniF52qFrX2qVmEfGo"`
}

func FromEnv() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("config.FromEnv: %w", err)
	}

	return cfg, nil
}
