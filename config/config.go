package config

import (
	"fmt"
	"time"

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
	Auth       Auth       `env-prefix:"AUTH_"`
}

type GRPCServer struct {
	Host                 string        `env:"HOST" env-default:"0.0.0.0"`
	Port                 int           `env:"PORT" env-default:"8082"`
	MaxConcurrentStreams uint32        `env:"MAX_CONCURRENT_STREAMS" env-default:"10"`
	KeepAliveTime        time.Duration `env:"KEEP_ALIVE_TIME" env-default:"10s"`
	KeepAliveTimeout     time.Duration `env:"KEEP_ALIVE_TIMEOUT" env-default:"10s"`
}

type Auth struct {
	Token string `env:"TOKEN" env-required:"true"`
}
