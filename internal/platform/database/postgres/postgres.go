// Package postgres provides PostgreSQL connection, migration, and metrics
// infrastructure. The composition root owns the *pgxpool.Pool; the functions
// here borrow it and never close it.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// DefaultPingTimeoutSeconds is the default timeout for database ping operations
	DefaultPingTimeoutSeconds = 15
)

// Ping verifies the pool can reach the database. A positive timeout bounds the
// attempt; zero or negative defers entirely to the caller's context.
func Ping(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
	if pool == nil {
		return fmt.Errorf("pool is nil")
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return pool.Ping(ctx)
}
