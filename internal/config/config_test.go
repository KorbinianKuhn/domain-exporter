package config

import (
	"os"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("config.yml", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGetReadsConfigFile(t *testing.T) {
	writeConfig(t, `
check_interval_in_seconds: 60
logging:
  level: debug
  format: json
domains:
  - example.com
  - münchen.de
`)

	c, err := Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.CheckIntervalInSeconds != 60 {
		t.Errorf("CheckIntervalInSeconds = %d, want 60", c.CheckIntervalInSeconds)
	}
	if c.Logging.Format != LogFormatJSON {
		t.Errorf("Logging.Format = %q, want %q", c.Logging.Format, LogFormatJSON)
	}
	if len(c.Domains) != 2 {
		t.Errorf("Domains = %v, want 2 entries", c.Domains)
	}
}

func TestGetDefaults(t *testing.T) {
	writeConfig(t, "domains:\n  - example.com\n")

	c, err := Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.CheckIntervalInSeconds != 86400 {
		t.Errorf("CheckIntervalInSeconds = %d, want 86400", c.CheckIntervalInSeconds)
	}
	if c.Logging.Format != LogFormatText {
		t.Errorf("Logging.Format = %q, want %q", c.Logging.Format, LogFormatText)
	}
}

func TestGetEnvOverridesConfigFile(t *testing.T) {
	writeConfig(t, "check_interval_in_seconds: 60\n")
	t.Setenv("CHECK_INTERVAL_IN_SECONDS", "120")
	t.Setenv("LOGGING_LEVEL", "warn")

	c, err := Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.CheckIntervalInSeconds != 120 {
		t.Errorf("CheckIntervalInSeconds = %d, want 120", c.CheckIntervalInSeconds)
	}
	if c.Logging.Level != "warn" {
		t.Errorf("Logging.Level = %q, want warn", c.Logging.Level)
	}
}

func TestGetValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{"invalid log level", "logging:\n  level: verbose\n", "invalid log level"},
		{"invalid log format", "logging:\n  format: xml\n", "invalid log format"},
		{"zero interval", "check_interval_in_seconds: 0\n", "invalid check interval"},
		{"negative interval", "check_interval_in_seconds: -1\n", "invalid check interval"},
		{"domain with scheme", "domains:\n  - https://example.com\n", "invalid domain format"},
		{"domain with space", "domains:\n  - exa mple.com\n", "invalid domain format"},
		{"domain without tld", "domains:\n  - localhost\n", "invalid domain format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeConfig(t, tt.config)

			_, err := Get()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
