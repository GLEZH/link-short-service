package config

import "testing"

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
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
			name:    "unknown flag",
			args:    []string{"-x", "test"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
