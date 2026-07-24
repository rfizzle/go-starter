package config

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TestUnmarshalResolvesNestedKeys guards the viper wiring that is the substance
// of the config layer: nested keys must resolve through flags, environment
// variables, and defaults into the mapstructure-tagged struct. It mirrors the
// cmd/serve.go setup (BindPFlags + AutomaticEnv + a "." / "-" -> "_" replacer)
// using an isolated viper instance so global state does not leak between tests.
func TestUnmarshalResolvesNestedKeys(t *testing.T) {
	// Not parallel: relies on t.Setenv, which forbids t.Parallel.
	v := viper.New()
	v.SetEnvPrefix("GOSTARTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	flags := pflag.NewFlagSet("serve", pflag.ContinueOnError)
	flags.String("environment", "production", "")
	flags.String("logging.level", "info", "")
	flags.String("logging.format", "", "")
	flags.Int("webserver.port", 8080, "")
	flags.String("database.uri", "", "")
	flags.Bool("database.auto-migrate", false, "")
	flags.String("seed.email", "", "")
	if err := v.BindPFlags(flags); err != nil {
		t.Fatalf("BindPFlags() error: %v", err)
	}

	// A nested env var must resolve through the "." -> "_" replacer.
	t.Setenv("GOSTARTER_LOGGING_LEVEL", "debug")
	t.Setenv("GOSTARTER_ENVIRONMENT", "development")
	t.Setenv("GOSTARTER_WEBSERVER_PORT", "9090")
	t.Setenv("GOSTARTER_DATABASE_AUTO_MIGRATE", "true")
	t.Setenv("GOSTARTER_SEED_EMAIL", "developer@example.com")

	var cfg Application
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want %q (from env)", cfg.Environment, "development")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q (from env)", cfg.Logging.Level, "debug")
	}
	if cfg.Webserver.Port != 9090 {
		t.Errorf("Webserver.Port = %d, want %d (from env)", cfg.Webserver.Port, 9090)
	}
	if !cfg.Database.AutoMigrate {
		t.Error("Database.AutoMigrate = false, want true (from env)")
	}
	if cfg.Seed.Email != "developer@example.com" {
		t.Errorf("Seed.Email = %q, want %q (from env)", cfg.Seed.Email, "developer@example.com")
	}
	// A key with only a flag default (no env/flag override) must fall back to it.
	if cfg.Logging.Format != "" {
		t.Errorf("Logging.Format = %q, want empty (flag default)", cfg.Logging.Format)
	}
}

// TestUnmarshalFlagBeatsEnv confirms the documented precedence: an explicitly
// set flag outranks an environment variable for the same key.
func TestUnmarshalFlagBeatsEnv(t *testing.T) {
	v := viper.New()
	v.SetEnvPrefix("GOSTARTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	flags := pflag.NewFlagSet("serve", pflag.ContinueOnError)
	flags.String("logging.level", "info", "")
	if err := flags.Set("logging.level", "error"); err != nil {
		t.Fatalf("flags.Set() error: %v", err)
	}
	if err := v.BindPFlags(flags); err != nil {
		t.Fatalf("BindPFlags() error: %v", err)
	}

	t.Setenv("GOSTARTER_LOGGING_LEVEL", "warn")

	var cfg Application
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if cfg.Logging.Level != "error" {
		t.Errorf("Logging.Level = %q, want %q (flag beats env)", cfg.Logging.Level, "error")
	}
}
