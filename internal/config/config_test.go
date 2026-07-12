package config

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		envAddress  string
		envBaseURL  string
		useAddress  bool
		useBaseURL  bool
		wantAddress string
		wantBaseURL string
		wantErr     bool
	}{
		{
			name:        "default values",
			args:        []string{},
			wantAddress: ":8080",
			wantBaseURL: "http://localhost:8080",
		},
		{
			name:        "custom values",
			args:        []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			wantAddress: "localhost:8888",
			wantBaseURL: "http://localhost:8000",
		},
		{
			name:        "env values have higher priority than flags",
			args:        []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envAddress:  "localhost:9999",
			envBaseURL:  "http://localhost:9000",
			useAddress:  true,
			useBaseURL:  true,
			wantAddress: "localhost:9999",
			wantBaseURL: "http://localhost:9000",
		},
		{
			name:        "server env overrides only server address",
			args:        []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envAddress:  "localhost:9999",
			useAddress:  true,
			wantAddress: "localhost:9999",
			wantBaseURL: "http://localhost:8000",
		},
		{
			name:        "base url env overrides only base url",
			args:        []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envBaseURL:  "http://localhost:9000",
			useBaseURL:  true,
			wantAddress: "localhost:8888",
			wantBaseURL: "http://localhost:9000",
		},
		{
			name:    "unknown flag",
			args:    []string{"-x", "test"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setEnv(t, "SERVER_ADDRESS", test.envAddress, test.useAddress)
			setEnv(t, "BASE_URL", test.envBaseURL, test.useBaseURL)

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
		})
	}
}

func setEnv(t *testing.T, key, value string, ok bool) {
	t.Helper()

	oldValue, existed := os.LookupEnv(key)
	if ok {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("os.Setenv() error = %v", err)
		}
	} else {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("os.Unsetenv() error = %v", err)
		}
	}

	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, oldValue)
			return
		}
		_ = os.Unsetenv(key)
	})
}
