package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

func LoadConfig() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return cfg, fmt.Errorf("read env: %w", err)
	}

	return cfg, nil
}

type Config struct {
	GRPCServer GRPCServer `env-prefix:"GRPC_SERVER_"`
}

type GRPCServer struct {
	Host string `env:"HOST" env-default:"0.0.0.0"`
	Port int    `env:"PORT" env-default:"8082"`
}
