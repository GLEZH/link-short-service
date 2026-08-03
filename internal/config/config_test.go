package config

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	oldAddress, hadAddress := os.LookupEnv("SERVER_ADDRESS")
	oldBaseURL, hadBaseURL := os.LookupEnv("BASE_URL")
	oldFilePath, hadFilePath := os.LookupEnv("FILE_STORAGE_PATH")
	oldDatabaseDSN, hadDatabaseDSN := os.LookupEnv("DATABASE_DSN")
	t.Cleanup(func() {
		restoreEnv("SERVER_ADDRESS", oldAddress, hadAddress)
		restoreEnv("BASE_URL", oldBaseURL, hadBaseURL)
		restoreEnv("FILE_STORAGE_PATH", oldFilePath, hadFilePath)
		restoreEnv("DATABASE_DSN", oldDatabaseDSN, hadDatabaseDSN)
	})

	tests := []struct {
		name            string
		args            []string
		envAddress      string
		envBaseURL      string
		envFilePath     string
		envDatabaseDSN  string
		useAddress      bool
		useBaseURL      bool
		useFilePath     bool
		useDatabaseDSN  bool
		wantAddress     string
		wantBaseURL     string
		wantFileStorage string
		wantDatabaseDSN string
		wantErr         bool
	}{
		{
			name:            "default values",
			args:            []string{},
			wantAddress:     ":8080",
			wantBaseURL:     "http://localhost:8080",
			wantFileStorage: "/tmp/short-url-db.json",
			wantDatabaseDSN: "",
		},
		{
			name:            "custom values",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000", "-f", "/tmp/flag.json", "-d", "postgres://flag"},
			wantAddress:     "localhost:8888",
			wantBaseURL:     "http://localhost:8000",
			wantFileStorage: "/tmp/flag.json",
			wantDatabaseDSN: "postgres://flag",
		},
		{
			name:            "env values have higher priority than flags",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000", "-f", "/tmp/flag.json", "-d", "postgres://flag"},
			envAddress:      "localhost:9999",
			envBaseURL:      "http://localhost:9000",
			envFilePath:     "/tmp/env.json",
			envDatabaseDSN:  "postgres://env",
			useAddress:      true,
			useBaseURL:      true,
			useFilePath:     true,
			useDatabaseDSN:  true,
			wantAddress:     "localhost:9999",
			wantBaseURL:     "http://localhost:9000",
			wantFileStorage: "/tmp/env.json",
			wantDatabaseDSN: "postgres://env",
		},
		{
			name:            "server env overrides only server address",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envAddress:      "localhost:9999",
			useAddress:      true,
			wantAddress:     "localhost:9999",
			wantBaseURL:     "http://localhost:8000",
			wantFileStorage: "/tmp/short-url-db.json",
			wantDatabaseDSN: "",
		},
		{
			name:            "base url env overrides only base url",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envBaseURL:      "http://localhost:9000",
			useBaseURL:      true,
			wantAddress:     "localhost:8888",
			wantBaseURL:     "http://localhost:9000",
			wantFileStorage: "/tmp/short-url-db.json",
			wantDatabaseDSN: "",
		},
		{
			name:    "unknown flag",
			args:    []string{"-x", "test"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = os.Unsetenv("SERVER_ADDRESS")
			_ = os.Unsetenv("BASE_URL")
			_ = os.Unsetenv("FILE_STORAGE_PATH")
			_ = os.Unsetenv("DATABASE_DSN")

			if test.useAddress {
				t.Setenv("SERVER_ADDRESS", test.envAddress)
			}
			if test.useBaseURL {
				t.Setenv("BASE_URL", test.envBaseURL)
			}
			if test.useFilePath {
				t.Setenv("FILE_STORAGE_PATH", test.envFilePath)
			}
			if test.useDatabaseDSN {
				t.Setenv("DATABASE_DSN", test.envDatabaseDSN)
			}

			cfg, err := New(test.args)
			if test.wantErr {
				if err == nil {
					t.Fatalf("New() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("New() unexpected error = %v", err)
			}

			if cfg.ServerAddress != test.wantAddress {
				t.Errorf("ServerAddress = %q, want %q", cfg.ServerAddress, test.wantAddress)
			}

			if cfg.BaseURL != test.wantBaseURL {
				t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, test.wantBaseURL)
			}

			if cfg.FileStoragePath != test.wantFileStorage {
				t.Errorf("FileStoragePath = %q, want %q", cfg.FileStoragePath, test.wantFileStorage)
			}

			if cfg.DatabaseDSN != test.wantDatabaseDSN {
				t.Errorf("DatabaseDSN = %q, want %q", cfg.DatabaseDSN, test.wantDatabaseDSN)
			}
		})
	}
}

func restoreEnv(key, value string, ok bool) {
	if ok {
		_ = os.Setenv(key, value)
		return
	}
	_ = os.Unsetenv(key)
}
