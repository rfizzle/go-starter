package observability

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rfizzle/go-starter/internal/platform/config"
)

// NewLogger builds a *zap.Logger from the application's logging and environment
// configuration. The level, encoding, and defaults are all driven by config so
// the same binary can emit human-readable console logs in development and
// structured JSON in production without code changes.
//
// This lives in the observability package so a single package owns the
// construction of all three telemetry signals — logger, tracer, and meter (see
// Init). Callers build the logger here and install the tracer/meter providers
// with Init.
//
// Precedence for the level: an explicit Logging.Level wins; otherwise it
// defaults to "info". Callers wanting verbose output should set Logging.Level to
// "debug" before calling. The resolved environment is attached as a field to
// every log line.
func NewLogger(cfg *config.Application) (*zap.Logger, error) {
	level, err := parseLevel(cfg.Logging.Level)
	if err != nil {
		return nil, err
	}

	// Build from the production preset in both cases and only vary the encoding,
	// so console vs. JSON is purely a formatting choice: it does not drag in the
	// development preset's side effects (DPanic actually panicking, stacktraces
	// emitted at Warn instead of Error). Console encoding is friendlier for local
	// development; JSON is what log aggregators expect everywhere else, and an
	// explicit Logging.Format overrides the environment-derived default.
	zapCfg := zap.NewProductionConfig()
	if resolveFormat(cfg.Logging.Format, cfg.Environment) == "console" {
		zapCfg.Encoding = "console"
		// Human-readable, colourised output for a terminal rather than the
		// production preset's epoch timestamps and plain levels.
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("building logger: %w", err)
	}

	if cfg.Environment != "" {
		logger = logger.With(zap.String("environment", cfg.Environment))
	}

	return logger, nil
}

// parseLevel maps a config string to a zapcore.Level, defaulting to info when
// unset and erroring on an unrecognised value so misconfiguration fails loudly.
func parseLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return zapcore.InfoLevel, nil
	case "debug":
		return zapcore.DebugLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("invalid log level %q (want debug, info, warn, or error)", level)
	}
}

// resolveFormat picks the encoder, honouring an explicit format and otherwise
// defaulting to console in development and JSON everywhere else.
func resolveFormat(format, environment string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "console":
		return "console"
	case "json":
		return "json"
	}

	if strings.EqualFold(strings.TrimSpace(environment), "development") {
		return "console"
	}
	return "json"
}
