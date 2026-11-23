package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	env "github.com/caarlos0/env/v11"
)

// Config contains the service settings
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	AuthSecret      string `env:"AUTH_SECRET" json:"auth_secret"`
	AuditFile       string `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL        string `env:"AUDIT_URL" json:"audit_url"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

// New parses settings from a JSON file, command-line arguments and the environment
func New(args []string) (*Config, error) {
	cfg := &Config{}
	var configPath string

	flags := flag.NewFlagSet("shortener", flag.ContinueOnError)
	flags.StringVar(&cfg.ServerAddress, "a", ":8080", "HTTP server address")
	flags.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base for short links")
	flags.StringVar(&cfg.FileStoragePath, "f", "", "path to file storage")
	flags.StringVar(&cfg.DatabaseDSN, "d", "", "database connection string")
	flags.StringVar(&cfg.AuthSecret, "auth-secret", "shortener-auth-secret", "auth secret")
	flags.StringVar(&cfg.AuditFile, "audit-file", "", "path to audit log file")
	flags.StringVar(&cfg.AuditURL, "audit-url", "", "audit server URL")
	flags.BoolVar(&cfg.EnableHTTPS, "s", false, "enable HTTPS")
	flags.StringVar(&configPath, "c", "", "path to JSON configuration file")
	flags.StringVar(&configPath, "config", "", "path to JSON configuration file")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if path := os.Getenv("CONFIG"); path != "" {
		configPath = path
	}
	if configPath != "" {
		if err := loadFile(configPath, cfg); err != nil {
			return nil, err
		}
		if err := flags.Parse(args); err != nil {
			return nil, err
		}
	}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}

	return nil
}
