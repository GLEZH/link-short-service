package config

import "flag"

type Config struct {
	ServerAddress string
	BaseURL       string
}

func New(args []string) (*Config, error) {
	cfg := &Config{}

	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&cfg.ServerAddress, "a", ":8080", "HTTP server address")
	flags.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base for short links")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return cfg, nil
}
