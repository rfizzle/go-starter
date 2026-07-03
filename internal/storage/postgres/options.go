package postgres

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Option configures a Client instance during initialization.
// Options are passed to the Open function to customize the client's behavior.
type Option func(p *Client) error

// WithURI configures the client to connect using a PostgreSQL connection URI.
// The context is used for connection timeout and cancellation.
// Example: "postgres://user:password@localhost/dbname" or "postgres://user:password@localhost:5432/dbname"
func WithURI(ctx context.Context, uri string) Option {
	return func(p *Client) error {
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

// WithConnection configures the client to use an existing PostgreSQL connection.
// Useful for sharing connections or when you need custom connection configuration.
func WithConnection(conn *pgx.Conn) Option {
	return func(p *Client) error {
		p.db = conn
		p.closeFunc = func(ctx context.Context) error {
			return conn.Close(ctx)
		}

		return nil
	}
}

// WithPool configures the client to use an existing PostgreSQL connection pool.
// This is the recommended option for production use as it provides connection pooling.
func WithPool(pool *pgxpool.Pool) Option {
	return func(p *Client) error {
		p.db = pool
		p.pool = pool
		p.closeFunc = func(ctx context.Context) error {
			pool.Close()
			return nil
		}

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
