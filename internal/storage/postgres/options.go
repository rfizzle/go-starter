package postgres

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
)

// Option configures a Provider instance during initialization.
// Options are passed to the Open function to customize the provider's behavior.
type Option func(p *Provider) error

// WithURI configures the provider to connect using a PostgreSQL connection URI.
// The context is used for connection timeout and cancellation.
// Example: "postgres://user:password@localhost/dbname" or "postgres://user:password@localhost:5432/dbname"
func WithURI(ctx context.Context, uri string) Option {
	return func(p *Provider) error {
		conn, err := pgxpool.New(ctx, uri)
		if err != nil {
			return err
		}

		p.db = conn
		p.pool = conn
		p.closeFunc = func(ctx context.Context) error {
			conn.Close()
			return nil
		}

		return nil
	}
}

// WithConnection configures the provider to use an existing PostgreSQL connection.
// Useful for sharing connections or when you need custom connection configuration.
func WithConnection(conn *pgx.Conn) Option {
	return func(p *Provider) error {
		p.db = conn
		p.closeFunc = func(ctx context.Context) error {
			return conn.Close(ctx)
		}

		return nil
	}
}

// WithPool configures the provider to use an existing PostgreSQL connection pool.
// This is the recommended option for production use as it provides connection pooling.
func WithPool(pool *pgxpool.Pool) Option {
	return func(p *Provider) error {
		p.db = pool
		p.pool = pool
		p.closeFunc = func(ctx context.Context) error {
			pool.Close()
			return nil
		}

		return nil
	}
}

// WithMockPool configures the provider to use a mock PostgreSQL pool for testing.
// This option should only be used in tests to simulate database behavior.
func WithMockPool(pool pgxmock.PgxPoolIface) Option {
	return func(p *Provider) error {
		p.db = pool
		p.closeFunc = func(ctx context.Context) error {
			return nil
		}

		return nil
	}
}

// WithPingTimeout sets the timeout duration in seconds for database ping operations.
// Default timeout is DefaultPingTimeoutSeconds. Set to 0 to disable timeout.
func WithPingTimeout(timeout int) Option {
	return func(p *Provider) error {
		p.pingTimeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// WithLogger configures the provider to use a specific logger instance.
// If not provided, operations will use a no-op logger.
func WithLogger(logger *zap.Logger) Option {
	return func(p *Provider) error {
		if logger == nil {
			return fmt.Errorf("logger cannot be nil")
		}

		p.logger = logger
		return nil
	}
}
