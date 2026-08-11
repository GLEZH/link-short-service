package config

import (
	"flag"

	env "github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	AuthSecret      string `env:"AUTH_SECRET"`
}

func New(args []string) (*Config, error) {
	cfg := &Config{}

	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&cfg.ServerAddress, "a", ":8080", "HTTP server address")
	flags.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base for short links")
	flags.StringVar(&cfg.FileStoragePath, "f", "", "path to file storage")
	flags.StringVar(&cfg.DatabaseDSN, "d", "", "database connection string")
	flags.StringVar(&cfg.AuthSecret, "auth-secret", "shortener-auth-secret", "auth secret")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
