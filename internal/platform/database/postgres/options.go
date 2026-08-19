package postgres

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Option configures a Client instance during initialization.
// Options are passed to the Open function to customize the client's behavior.
type Option func(p *Client) error

// WithPool configures the client to use an existing PostgreSQL connection pool.
// The caller owns the pool and remains responsible for closing it.
func WithPool(pool *pgxpool.Pool) Option {
	return func(p *Client) error {
		p.db = pool
		p.pool = pool
		return nil
	}
}

// WithPingTimeout sets the timeout duration in seconds for database ping operations.
// Default timeout is DefaultPingTimeoutSeconds. Set to 0 to disable timeout.
func WithPingTimeout(timeout int) Option {
	return func(p *Client) error {
		p.pingTimeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// WithLogger configures the client to use a specific logger instance.
// If not provided, operations will use a no-op logger.
func WithLogger(logger *zap.Logger) Option {
	return func(p *Client) error {
		if logger == nil {
			return fmt.Errorf("logger cannot be nil")
		}

		p.logger = logger
		return nil
	}
}
