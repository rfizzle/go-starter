package config

import (
	"math"
	"strings"
	"testing"
)

func TestApplicationValidate(t *testing.T) {
	tests := []struct {
		name    string
		scope   ValidationScope
		mutate  func(*Application)
		wantErr []string
	}{
		{
			name:  "valid serve configuration",
			scope: ValidationScopeServe,
		},
		{
			name:  "valid seed configuration ignores unused serve settings",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Environment = "development"
				cfg.Webserver.Port = -1
				cfg.Auth.JWTSecret = ""
				cfg.Auth.JWTTTL = -1
				cfg.Observability.Endpoint = "https://collector.example.com:4317"
				cfg.Observability.SampleRatio = 2
				cfg.Database.CloseTimeout = -1
			},
		},
		{
			name:  "seed command is development only",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Environment = "production"
			},
			wantErr: []string{"only supported in the development environment"},
		},
		{
			name:  "seed email is required",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Environment = "development"
				cfg.Seed.Email = ""
			},
			wantErr: []string{"seed.email"},
		},
		{
			name:  "seed password is required",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Environment = "development"
				cfg.Seed.Password = ""
			},
			wantErr: []string{"seed.password is required"},
		},
		{
			name:  "seed password must fit bcrypt",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Environment = "development"
				cfg.Seed.Password = strings.Repeat("p", 73)
			},
			wantErr: []string{"seed.password"},
		},
		{
			name:  "valid migrate configuration ignores unused serve settings",
			scope: ValidationScopeMigrate,
			mutate: func(cfg *Application) {
				cfg.Webserver.Port = -1
				cfg.Auth.JWTSecret = ""
				cfg.Auth.JWTTTL = -1
				cfg.Observability.Endpoint = "https://collector.example.com:4317"
				cfg.Observability.SampleRatio = 2
				cfg.Database.CloseTimeout = -1
			},
		},
		{
			name:  "environment is restricted",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Environment = "local"
			},
			wantErr: []string{"environment"},
		},
		{
			name:  "logging values are restricted",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Logging.Level = "trace"
				cfg.Logging.Format = "text"
			},
			wantErr: []string{"logging.level", "logging.format"},
		},
		{
			name:  "database URI is required",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Database.URI = ""
			},
			wantErr: []string{"database.uri"},
		},
		{
			name:  "database URI must be PostgreSQL",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Database.URI = "mysql://user:password@localhost/example"
			},
			wantErr: []string{"database.uri"},
		},
		{
			name:  "database durations and sizes cannot be negative",
			scope: ValidationScopeSeed,
			mutate: func(cfg *Application) {
				cfg.Database.PingTimeout = -1
				cfg.Database.MaxConns = -1
				cfg.Database.MinConns = -1
				cfg.Database.MaxConnLifetime = -1
				cfg.Database.MaxConnIdleTime = -1
				cfg.Database.HealthCheckPeriod = -1
			},
			wantErr: []string{
				"database.ping-timeout",
				"database.max-conns",
				"database.min-conns",
				"database.max-conn-lifetime",
				"database.max-conn-idle-time",
				"database.health-check-period",
			},
		},
		{
			name:  "database close timeout cannot be negative for serve",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Database.CloseTimeout = -1
			},
			wantErr: []string{"database.close-timeout"},
		},
		{
			name:  "database minimum cannot exceed maximum",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Database.MinConns = 6
				cfg.Database.MaxConns = 5
			},
			wantErr: []string{"database.min-conns"},
		},
		{
			name:  "webserver port must be in range",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Webserver.Port = 65536
			},
			wantErr: []string{"webserver.port"},
		},
		{
			name:  "webserver durations and sizes cannot be negative",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.ShutdownDelay = -1
				cfg.Webserver.ReadHeaderTimeout = -1
				cfg.Webserver.ReadTimeout = -1
				cfg.Webserver.WriteTimeout = -1
				cfg.Webserver.IdleTimeout = -1
				cfg.Webserver.RequestTimeout = -1
				cfg.Webserver.ShutdownTimeout = -1
				cfg.Webserver.MaxHeaderBytes = -1
				cfg.Webserver.MaxBodyBytes = -1
			},
			wantErr: []string{
				"shutdown-delay",
				"webserver.read-header-timeout",
				"webserver.read-timeout",
				"webserver.write-timeout",
				"webserver.idle-timeout",
				"webserver.request-timeout",
				"webserver.shutdown-timeout",
				"webserver.max-header-bytes",
				"webserver.max-body-bytes",
			},
		},
		{
			name:  "TLS certificate and key must be paired",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Webserver.TLSCertFile = "server.crt"
				cfg.Webserver.TLSKeyFile = ""
			},
			wantErr: []string{"webserver.tls-cert-file", "webserver.tls-key-file"},
		},
		{
			name:  "JWT policy requires a strong secret and positive TTL",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Auth.JWTSecret = "short"
				cfg.Auth.JWTTTL = 0
			},
			wantErr: []string{"auth.jwt-secret", "auth.jwt-ttl"},
		},
		{
			name:  "rate limit bounds must be positive",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.RateLimit.MaxRequests = 0
				cfg.RateLimit.Window = -1
				cfg.RateLimit.MaxKeys = 0
				cfg.RateLimit.CleanupInterval = -1
			},
			wantErr: []string{
				"rate-limit.max-requests",
				"rate-limit.window",
				"rate-limit.max-keys",
				"rate-limit.cleanup-interval",
			},
		},
		{
			name:  "trusted proxies must be canonical CIDR networks",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.RateLimit.TrustedProxies = []string{
					"192.0.2.1",
					"192.0.2.1/24",
					"not-a-network",
				}
			},
			wantErr: []string{
				"rate-limit.trusted-proxies[0]",
				"rate-limit.trusted-proxies[1]",
				"rate-limit.trusted-proxies[2]",
			},
		},
		{
			name:  "enabled observability validates endpoint and ratio",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Observability.Endpoint = "https://collector.example.com:4317"
				cfg.Observability.SampleRatio = math.NaN()
			},
			wantErr: []string{"observability.endpoint", "observability.sample-ratio"},
		},
		{
			name:  "multiple independent errors are returned together",
			scope: ValidationScopeServe,
			mutate: func(cfg *Application) {
				cfg.Environment = "invalid"
				cfg.Database.URI = ""
				cfg.Webserver.Port = 0
				cfg.Auth.JWTSecret = ""
			},
			wantErr: []string{"environment", "database.uri", "webserver.port", "auth.jwt-secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validApplication()
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}

			err := cfg.Validate(tt.scope)
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() error = nil, want validation error")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Validate() error = %q, want field %q", err, want)
				}
			}
		})
	}
}

func TestApplicationValidate_DoesNotExposeSecrets(t *testing.T) {
	databaseValue := strings.Join([]string{"database", "private", "value"}, "-")
	jwtValue := strings.Join([]string{"jwt", "private", "value"}, "-")

	cfg := validApplication()
	cfg.Database.URI = "mysql://user:" + databaseValue + "@localhost/example"
	cfg.Auth.JWTSecret = jwtValue

	err := cfg.Validate(ValidationScopeServe)
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
	for _, sensitiveValue := range []string{databaseValue, jwtValue} {
		if strings.Contains(err.Error(), sensitiveValue) {
			t.Errorf("Validate() error exposed sensitive value %q", sensitiveValue)
		}
	}
}

func TestApplicationValidate_RejectsUnknownScope(t *testing.T) {
	err := validApplication().Validate("unknown")
	if err == nil || !strings.Contains(err.Error(), "unsupported configuration validation scope") {
		t.Fatalf("Validate() error = %v, want unsupported scope error", err)
	}
}

func validApplication() Application {
	return Application{
		Environment: "production",
		Webserver: Webserver{
			Port: 8080,
		},
		Database: Database{
			URI:          "postgres://localhost:5432/example",
			PingTimeout:  15,
			MaxConns:     10,
			MinConns:     2,
			CloseTimeout: 10,
		},
		Seed: Seed{
			Email:    "developer@example.com",
			Password: "development-password",
		},
		Auth: Auth{
			JWTSecret: strings.Repeat("s", 32),
			JWTTTL:    3600,
		},
		RateLimit: RateLimit{
			MaxRequests:     5,
			Window:          60,
			MaxKeys:         10_000,
			CleanupInterval: 60,
			TrustedProxies:  []string{"10.0.0.0/8", "2001:db8::/32"},
		},
		Logging: Logging{
			Level:  "info",
			Format: "json",
		},
		Observability: Observability{
			Enabled:     true,
			Endpoint:    "localhost:4317",
			SampleRatio: 1,
		},
	}
}
