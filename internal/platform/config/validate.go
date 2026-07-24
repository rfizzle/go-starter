package config

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ValidationScope identifies the command whose configuration is being
// validated. Commands validate only the settings they consume.
type ValidationScope string

const (
	ValidationScopeServe   ValidationScope = "serve"
	ValidationScopeSeed    ValidationScope = "seed"
	ValidationScopeMigrate ValidationScope = "migrate"
)

// Validate checks the application settings used by scope and returns all
// configuration problems together. Error messages name fields and policies,
// but never include configuration values because some fields contain secrets.
func (application Application) Validate(scope ValidationScope) error {
	var errs []error

	switch scope {
	case ValidationScopeServe, ValidationScopeSeed, ValidationScopeMigrate:
		errs = append(errs, application.validateShared(scope)...)
	default:
		return fmt.Errorf("unsupported configuration validation scope %q", scope)
	}

	if scope == ValidationScopeServe {
		errs = append(errs, application.validateServe()...)
	}
	if scope == ValidationScopeSeed {
		errs = append(errs, application.validateSeed()...)
	}

	return errors.Join(errs...)
}

func (application Application) validateShared(scope ValidationScope) []error {
	var errs []error

	if !oneOf(application.Environment, "development", "staging", "production") {
		errs = append(errs, errors.New("environment must be one of development, staging, or production"))
	}

	if !oneOfOrEmpty(application.Logging.Level, "debug", "info", "warn", "warning", "error") {
		errs = append(errs, errors.New("logging.level must be one of debug, info, warn, or error"))
	}
	if !oneOfOrEmpty(application.Logging.Format, "json", "console") {
		errs = append(errs, errors.New("logging.format must be one of json or console"))
	}

	errs = append(errs, validateDatabase(application.Database, scope == ValidationScopeServe)...)
	return errs
}

func (application Application) validateSeed() []error {
	var errs []error

	if !oneOf(application.Environment, "development") {
		errs = append(errs, errors.New("seed command is only supported in the development environment"))
	}
	if strings.TrimSpace(application.Seed.Email) == "" {
		errs = append(errs, errors.New("seed.email is required"))
	}
	if len(application.Seed.Password) > 72 {
		errs = append(errs, errors.New("seed.password must not exceed 72 bytes"))
	}

	return errs
}

func (application Application) validateServe() []error {
	var errs []error

	if application.Webserver.Port < 1 || application.Webserver.Port > 65535 {
		errs = append(errs, errors.New("webserver.port must be between 1 and 65535"))
	}

	nonnegativeInts := []struct {
		name  string
		value int
	}{
		{name: "shutdown-delay", value: application.ShutdownDelay},
		{name: "webserver.read-header-timeout", value: application.Webserver.ReadHeaderTimeout},
		{name: "webserver.read-timeout", value: application.Webserver.ReadTimeout},
		{name: "webserver.write-timeout", value: application.Webserver.WriteTimeout},
		{name: "webserver.idle-timeout", value: application.Webserver.IdleTimeout},
		{name: "webserver.request-timeout", value: application.Webserver.RequestTimeout},
		{name: "webserver.shutdown-timeout", value: application.Webserver.ShutdownTimeout},
		{name: "webserver.max-header-bytes", value: application.Webserver.MaxHeaderBytes},
	}
	for _, field := range nonnegativeInts {
		if field.value < 0 {
			errs = append(errs, fmt.Errorf("%s must not be negative", field.name))
		}
	}
	if application.Webserver.MaxBodyBytes < 0 {
		errs = append(errs, errors.New("webserver.max-body-bytes must not be negative"))
	}

	hasTLSCert := strings.TrimSpace(application.Webserver.TLSCertFile) != ""
	hasTLSKey := strings.TrimSpace(application.Webserver.TLSKeyFile) != ""
	if hasTLSCert != hasTLSKey {
		errs = append(errs, errors.New("webserver.tls-cert-file and webserver.tls-key-file must be set together"))
	}

	if len(application.Auth.JWTSecret) < 32 {
		errs = append(errs, errors.New("auth.jwt-secret must be at least 32 bytes"))
	}
	if application.Auth.JWTTTL <= 0 {
		errs = append(errs, errors.New("auth.jwt-ttl must be greater than zero"))
	}

	if application.Observability.Enabled {
		ratio := application.Observability.SampleRatio
		if math.IsNaN(ratio) || math.IsInf(ratio, 0) || ratio < 0 || ratio > 1 {
			errs = append(errs, errors.New("observability.sample-ratio must be between 0 and 1"))
		}
		if endpoint := strings.TrimSpace(application.Observability.Endpoint); endpoint != "" && !validOTLPEndpoint(endpoint) {
			errs = append(errs, errors.New("observability.endpoint must be a host:port address without a URL scheme"))
		}
	}

	return errs
}

func validateDatabase(database Database, validateCloseTimeout bool) []error {
	var errs []error

	if !validDatabaseURI(database.URI) {
		errs = append(errs, errors.New("database.uri must be a valid PostgreSQL connection URI"))
	}

	nonnegativeInts := []struct {
		name  string
		value int
	}{
		{name: "database.ping-timeout", value: database.PingTimeout},
		{name: "database.max-conns", value: database.MaxConns},
		{name: "database.min-conns", value: database.MinConns},
		{name: "database.max-conn-lifetime", value: database.MaxConnLifetime},
		{name: "database.max-conn-idle-time", value: database.MaxConnIdleTime},
		{name: "database.health-check-period", value: database.HealthCheckPeriod},
	}
	if validateCloseTimeout {
		nonnegativeInts = append(nonnegativeInts, struct {
			name  string
			value int
		}{name: "database.close-timeout", value: database.CloseTimeout})
	}
	for _, field := range nonnegativeInts {
		if field.value < 0 {
			errs = append(errs, fmt.Errorf("%s must not be negative", field.name))
		}
	}

	if database.MaxConns > 0 && database.MinConns > database.MaxConns {
		errs = append(errs, errors.New("database.min-conns must not exceed database.max-conns"))
	}

	return errs
}

func validDatabaseURI(rawURI string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return false
	}

	return (parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") &&
		parsed.Host != "" &&
		strings.Trim(parsed.Path, "/") != ""
}

func validOTLPEndpoint(endpoint string) bool {
	if strings.Contains(endpoint, "://") {
		return false
	}

	host, rawPort, err := net.SplitHostPort(endpoint)
	if err != nil || strings.TrimSpace(host) == "" {
		return false
	}

	port, err := strconv.Atoi(rawPort)
	return err == nil && port >= 1 && port <= 65535
}

func oneOf(value string, allowed ...string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func oneOfOrEmpty(value string, allowed ...string) bool {
	return strings.TrimSpace(value) == "" || oneOf(value, allowed...)
}
