package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `env:"ENV" envDefault:"dev"`
	GRPC        GRPCConfig
	DatabaseUrl string        `env:"DATABASE_URL" env-required:"true"`
	TokenTTL    time.Duration `env:"TOKEN_TTL" envDefault:"1h"`
}

type GRPCConfig struct {
	Port    int           `env:"GRPC_PORT"`
	Timeout time.Duration `env:"GRPC_TIMEOUT"`
}

func MustLoad() *Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
