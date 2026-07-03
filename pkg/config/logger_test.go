package config

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    zapcore.Level
		wantErr bool
	}{
		{name: "empty defaults to info", input: "", want: zapcore.InfoLevel},
		{name: "info", input: "info", want: zapcore.InfoLevel},
		{name: "debug", input: "debug", want: zapcore.DebugLevel},
		{name: "warn", input: "warn", want: zapcore.WarnLevel},
		{name: "warning alias", input: "warning", want: zapcore.WarnLevel},
		{name: "error", input: "error", want: zapcore.ErrorLevel},
		{name: "case insensitive", input: "DEBUG", want: zapcore.DebugLevel},
		{name: "trims whitespace", input: "  info  ", want: zapcore.InfoLevel},
		{name: "invalid errors", input: "trace", want: zapcore.InfoLevel, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		format      string
		environment string
		want        string
	}{
		{name: "explicit console", format: "console", environment: "production", want: "console"},
		{name: "explicit json", format: "json", environment: "development", want: "json"},
		{name: "development defaults to console", format: "", environment: "development", want: "console"},
		{name: "production defaults to json", format: "", environment: "production", want: "json"},
		{name: "unknown env defaults to json", format: "", environment: "", want: "json"},
		{name: "explicit format wins over env", format: "json", environment: "development", want: "json"},
		{name: "case insensitive env", format: "", environment: "Development", want: "console"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolveFormat(tt.format, tt.environment); got != tt.want {
				t.Errorf("resolveFormat(%q, %q) = %q, want %q", tt.format, tt.environment, got, tt.want)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("builds with valid config", func(t *testing.T) {
		t.Parallel()

		logger, err := NewLogger(&Application{
			Environment: "production",
			Logging:     Logging{Level: "debug", Format: "json"},
		})
		if err != nil {
			t.Fatalf("NewLogger() unexpected error: %v", err)
		}
		if logger == nil {
			t.Fatal("NewLogger() returned nil logger")
		}
	})

	t.Run("rejects invalid level", func(t *testing.T) {
		t.Parallel()

		if _, err := NewLogger(&Application{Logging: Logging{Level: "bogus"}}); err == nil {
			t.Fatal("NewLogger() expected error for invalid level, got nil")
		}
	})
}
