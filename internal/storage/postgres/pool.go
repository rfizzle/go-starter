package postgres

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig carries the connection-pool tuning knobs applied on top of the
// settings parsed from the connection URI. Every field is optional: a zero
// value leaves the corresponding pgx default in place, so callers only set the
// knobs they care about. Counts are plain integers; durations are expressed in
// seconds to match the surrounding config surface.
type PoolConfig struct {
	// MaxConns caps the pool size. 0 keeps the pgx default (the greater of 4
	// and runtime.NumCPU()).
	MaxConns int
	// MinConns is the number of idle connections the pool keeps warm. 0 keeps
	// the pgx default (0).
	MinConns int
	// MaxConnLifetimeSeconds closes and replaces a connection this long after
	// it was created. 0 keeps the pgx default (1 hour).
	MaxConnLifetimeSeconds int
	// MaxConnIdleTimeSeconds closes a connection that has been idle this long.
	// 0 keeps the pgx default (30 minutes).
	MaxConnIdleTimeSeconds int
	// HealthCheckPeriodSeconds is the interval between pool health checks of
	// idle connections. 0 keeps the pgx default (1 minute).
	HealthCheckPeriodSeconds int
}

// NewPool parses the connection URI and returns a pgx pool configured with the
// supplied tuning knobs. Building the pool through ParseConfig + NewWithConfig
// (rather than pgxpool.New) is what lets callers size the pool explicitly
// instead of inheriting the all-default settings, which are frequently wrong
// under load. The pool connects lazily; callers should Ping before serving to
// surface an unreachable database at startup.
func NewPool(ctx context.Context, uri string, poolCfg PoolConfig) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig(): %w", err)
	}

	if poolCfg.MaxConns > 0 {
		cfg.MaxConns = clampInt32(poolCfg.MaxConns)
	}
	if poolCfg.MinConns > 0 {
		cfg.MinConns = clampInt32(poolCfg.MinConns)
	}
	if poolCfg.MaxConnLifetimeSeconds > 0 {
		cfg.MaxConnLifetime = time.Duration(poolCfg.MaxConnLifetimeSeconds) * time.Second
	}
	if poolCfg.MaxConnIdleTimeSeconds > 0 {
		cfg.MaxConnIdleTime = time.Duration(poolCfg.MaxConnIdleTimeSeconds) * time.Second
	}
	if poolCfg.HealthCheckPeriodSeconds > 0 {
		cfg.HealthCheckPeriod = time.Duration(poolCfg.HealthCheckPeriodSeconds) * time.Second
	}

	// NewWithConfig does not reject MinConns > MaxConns; catch it here so the
	// misconfiguration fails fast at startup rather than surfacing later as
	// odd pool behaviour.
	if cfg.MinConns > cfg.MaxConns {
		return nil, fmt.Errorf("database pool min-conns (%d) exceeds max-conns (%d)", cfg.MinConns, cfg.MaxConns)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig(): %w", err)
	}

	return pool, nil
}

// ClosePool closes the pool, bounding the wait by timeout. pgxpool.Pool.Close
// blocks until every connection is returned to the pool and closed, so a
// connection wedged mid-query (or on a dead socket) can otherwise hang shutdown
// indefinitely. On timeout ClosePool stops waiting and returns an error; the
// abandoned Close finishes on its own if it can make progress, and the process
// is exiting regardless. It returns nil once Close completes within the
// timeout. A non-positive timeout waits indefinitely (the raw Close behaviour).
func ClosePool(pool *pgxpool.Pool, timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()

	if timeout <= 0 {
		<-done
		return nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-timer.C:
		return fmt.Errorf("pgxpool.Close() did not complete within %s", timeout)
	}
}

// clampInt32 narrows a non-negative int to int32, capping at math.MaxInt32 so
// an out-of-range config value saturates rather than wrapping to a negative
// pool size.
func clampInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	// #nosec G115 -- v is bounded above by math.MaxInt32 (checked above) and
	// callers only pass non-negative counts, so the conversion cannot overflow.
	return int32(v)
}
