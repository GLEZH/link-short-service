package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	oldAddress, hadAddress := os.LookupEnv("SERVER_ADDRESS")
	oldBaseURL, hadBaseURL := os.LookupEnv("BASE_URL")
	oldFilePath, hadFilePath := os.LookupEnv("FILE_STORAGE_PATH")
	oldDatabaseDSN, hadDatabaseDSN := os.LookupEnv("DATABASE_DSN")
	oldAuthSecret, hadAuthSecret := os.LookupEnv("AUTH_SECRET")
	t.Cleanup(func() {
		restoreEnv("SERVER_ADDRESS", oldAddress, hadAddress)
		restoreEnv("BASE_URL", oldBaseURL, hadBaseURL)
		restoreEnv("FILE_STORAGE_PATH", oldFilePath, hadFilePath)
		restoreEnv("DATABASE_DSN", oldDatabaseDSN, hadDatabaseDSN)
		restoreEnv("AUTH_SECRET", oldAuthSecret, hadAuthSecret)
	})

	tests := []struct {
		name            string
		args            []string
		envAddress      string
		envBaseURL      string
		envFilePath     string
		envDatabaseDSN  string
		envAuthSecret   string
		useAddress      bool
		useBaseURL      bool
		useFilePath     bool
		useDatabaseDSN  bool
		useAuthSecret   bool
		wantAddress     string
		wantBaseURL     string
		wantFileStorage string
		wantDatabaseDSN string
		wantAuthSecret  string
		wantErr         bool
	}{
		{
			name:            "default values",
			args:            []string{},
			wantAddress:     ":8080",
			wantBaseURL:     "http://localhost:8080",
			wantFileStorage: "",
			wantDatabaseDSN: "",
			wantAuthSecret:  "shortener-auth-secret",
		},
		{
			name:            "custom values",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000", "-f", "/tmp/flag.json", "-d", "postgres://flag", "-auth-secret", "flag-secret"},
			wantAddress:     "localhost:8888",
			wantBaseURL:     "http://localhost:8000",
			wantFileStorage: "/tmp/flag.json",
			wantDatabaseDSN: "postgres://flag",
			wantAuthSecret:  "flag-secret",
		},
		{
			name:            "env values have higher priority than flags",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000", "-f", "/tmp/flag.json", "-d", "postgres://flag", "-auth-secret", "flag-secret"},
			envAddress:      "localhost:9999",
			envBaseURL:      "http://localhost:9000",
			envFilePath:     "/tmp/env.json",
			envDatabaseDSN:  "postgres://env",
			envAuthSecret:   "env-secret",
			useAddress:      true,
			useBaseURL:      true,
			useFilePath:     true,
			useDatabaseDSN:  true,
			useAuthSecret:   true,
			wantAddress:     "localhost:9999",
			wantBaseURL:     "http://localhost:9000",
			wantFileStorage: "/tmp/env.json",
			wantDatabaseDSN: "postgres://env",
			wantAuthSecret:  "env-secret",
		},
		{
			name:            "server env overrides only server address",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envAddress:      "localhost:9999",
			useAddress:      true,
			wantAddress:     "localhost:9999",
			wantBaseURL:     "http://localhost:8000",
			wantFileStorage: "",
			wantDatabaseDSN: "",
			wantAuthSecret:  "shortener-auth-secret",
		},
		{
			name:            "base url env overrides only base url",
			args:            []string{"-a", "localhost:8888", "-b", "http://localhost:8000"},
			envBaseURL:      "http://localhost:9000",
			useBaseURL:      true,
			wantAddress:     "localhost:8888",
			wantBaseURL:     "http://localhost:9000",
			wantFileStorage: "",
			wantDatabaseDSN: "",
			wantAuthSecret:  "shortener-auth-secret",
		},
		{
			name:            "file storage env",
			args:            []string{},
			envFilePath:     "/tmp/env.json",
			useFilePath:     true,
			wantAddress:     ":8080",
			wantBaseURL:     "http://localhost:8080",
			wantFileStorage: "/tmp/env.json",
			wantDatabaseDSN: "",
			wantAuthSecret:  "shortener-auth-secret",
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
			_ = os.Unsetenv("AUTH_SECRET")

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
			if test.useAuthSecret {
				t.Setenv("AUTH_SECRET", test.envAuthSecret)
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

			if cfg.AuthSecret != test.wantAuthSecret {
				t.Errorf("AuthSecret = %q, want %q", cfg.AuthSecret, test.wantAuthSecret)
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

func TestNewAuditConfig(t *testing.T) {
	oldAuditFile, hadAuditFile := os.LookupEnv("AUDIT_FILE")
	oldAuditURL, hadAuditURL := os.LookupEnv("AUDIT_URL")
	t.Cleanup(func() {
		restoreEnv("AUDIT_FILE", oldAuditFile, hadAuditFile)
		restoreEnv("AUDIT_URL", oldAuditURL, hadAuditURL)
	})

	_ = os.Unsetenv("AUDIT_FILE")
	_ = os.Unsetenv("AUDIT_URL")

	cfg, err := New([]string{"--audit-file", "/tmp/flag-audit.log", "--audit-url", "http://flag.example.com"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if cfg.AuditFile != "/tmp/flag-audit.log" {
		t.Errorf("AuditFile = %q, want %q", cfg.AuditFile, "/tmp/flag-audit.log")
	}
	if cfg.AuditURL != "http://flag.example.com" {
		t.Errorf("AuditURL = %q, want %q", cfg.AuditURL, "http://flag.example.com")
	}

	t.Setenv("AUDIT_FILE", "/tmp/env-audit.log")
	t.Setenv("AUDIT_URL", "http://env.example.com")
	cfg, err = New([]string{"--audit-file", "/tmp/flag-audit.log", "--audit-url", "http://flag.example.com"})
	if err != nil {
		t.Fatalf("New() with environment error = %v", err)
	}
	if cfg.AuditFile != "/tmp/env-audit.log" {
		t.Errorf("AuditFile = %q, want %q", cfg.AuditFile, "/tmp/env-audit.log")
	}
	if cfg.AuditURL != "http://env.example.com" {
		t.Errorf("AuditURL = %q, want %q", cfg.AuditURL, "http://env.example.com")
	}
}

func TestNewHTTPSConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     string
		want    bool
		wantErr bool
	}{
		{name: "disabled by default"},
		{name: "enabled by flag", args: []string{"-s"}, want: true},
		{name: "enabled by environment", env: "true", want: true},
		{name: "environment overrides flag", args: []string{"-s"}, env: "false"},
		{name: "environment enables explicit false flag", args: []string{"-s=false"}, env: "true", want: true},
		{name: "invalid environment", env: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENABLE_HTTPS", tt.env)
			cfg, err := New(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("New() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if cfg.EnableHTTPS != tt.want {
				t.Errorf("EnableHTTPS = %v, want %v", cfg.EnableHTTPS, tt.want)
			}
		})
	}
}

func TestNewFileConfig(t *testing.T) {
	for _, key := range []string{"CONFIG", "SERVER_ADDRESS", "BASE_URL", "FILE_STORAGE_PATH", "DATABASE_DSN", "AUTH_SECRET", "AUDIT_FILE", "AUDIT_URL", "ENABLE_HTTPS"} {
		t.Setenv(key, "")
	}
	defaults := Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080", AuthSecret: "shortener-auth-secret"}
	fileValues := Config{
		ServerAddress: "localhost:8443", BaseURL: "https://localhost:8443",
		FileStoragePath: "/tmp/urls.json", DatabaseDSN: "postgres://file",
		AuthSecret: "file-secret", AuditFile: "/tmp/audit.log",
		AuditURL: "https://audit.example.com", EnableHTTPS: true,
	}
	const fullJSON = `{"server_address":"localhost:8443","base_url":"https://localhost:8443","file_storage_path":"/tmp/urls.json","database_dsn":"postgres://file","auth_secret":"file-secret","audit_file":"/tmp/audit.log","audit_url":"https://audit.example.com","enable_https":true}`
	tests := []struct {
		name    string
		data    string
		flag    string
		args    []string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{name: "short flag", data: fullJSON, flag: "-c", want: fileValues},
		{name: "long flag", data: fullJSON, flag: "-config", want: fileValues},
		{name: "CONFIG selects file", data: fullJSON, want: fileValues},
		{name: "CONFIG overrides flag path", data: fullJSON, args: []string{"-c", "missing.json"}, want: fileValues},
		{name: "missing fields preserve defaults", data: `{}`, flag: "-c", want: defaults},
		{name: "partial file", data: `{"enable_https":true}`, flag: "-c", want: Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080", AuthSecret: "shortener-auth-secret", EnableHTTPS: true}},
		{name: "explicit empty values", data: `{"server_address":"","base_url":"","auth_secret":""}`, flag: "-c", want: Config{}},
		{name: "flags override all file settings", data: fullJSON, flag: "-config",
			args: []string{"-a", ":8080", "-b", "http://localhost:8080", "-f=", "-d=", "-auth-secret", "shortener-auth-secret", "-audit-file=", "-audit-url=", "-s=false"}, want: defaults},
		{name: "environment overrides flags and file", data: fullJSON, flag: "-c",
			args: []string{"-a", ":9999", "-s"},
			env:  map[string]string{"SERVER_ADDRESS": ":9000", "BASE_URL": "http://env", "FILE_STORAGE_PATH": "env.json", "DATABASE_DSN": "postgres://env", "AUTH_SECRET": "env-secret", "AUDIT_FILE": "env.log", "AUDIT_URL": "http://audit.env", "ENABLE_HTTPS": "false"},
			want: Config{ServerAddress: ":9000", BaseURL: "http://env", FileStoragePath: "env.json", DatabaseDSN: "postgres://env", AuthSecret: "env-secret", AuditFile: "env.log", AuditURL: "http://audit.env"}},
		{name: "invalid JSON", data: `{`, flag: "-c", wantErr: true},
		{name: "incorrect field type", data: `{"enable_https":"true"}`, flag: "-c", wantErr: true},
		{name: "trailing JSON", data: `{} {}`, flag: "-c", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tt.data), 0600); err != nil {
				t.Fatal(err)
			}
			args := append([]string{}, tt.args...)
			if tt.flag != "" {
				args = append(args, tt.flag, path)
			} else {
				t.Setenv("CONFIG", path)
			}
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			cfg, err := New(args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("New() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if !reflect.DeepEqual(*cfg, tt.want) {
				t.Errorf("config = %+v, want %+v", *cfg, tt.want)
			}
		})
	}
	t.Run("missing file", func(t *testing.T) {
		if _, err := New([]string{"-c", filepath.Join(t.TempDir(), "missing.json")}); err == nil {
			t.Fatal("New() error = nil, want error")
		}
	})
}
