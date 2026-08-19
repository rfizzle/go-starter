// Package postgres provides PostgreSQL connection, migration, and metrics
// infrastructure.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

const (
	// DefaultPingTimeoutSeconds is the default timeout for database ping operations
	DefaultPingTimeoutSeconds = 15
)

// pinger is the only database behavior the lifecycle client consumes.
// *pgx.Conn, *pgxpool.Pool, and pgxmock pools satisfy it structurally.
type pinger interface {
	Ping(ctx context.Context) error
}

// Client provides the PostgreSQL lifecycle operations used by the application.
// Domain queries live in repositories rather than on this type.
type Client struct {
	logger *zap.Logger
	db     pinger
	// pool is the underlying pgx connection pool when the client was built
	// with WithPool. It is nil for mock-backed clients. Migrate needs a real
	// pool because goose runs through a database/sql handle wrapped over it
	// (see migration.go).
	pool        *pgxpool.Pool
	pingTimeout time.Duration
}

// Open creates and initializes a new PostgreSQL database client.
// It accepts a context for timeout control and various configuration options.
// The client must be configured with a database connection using WithPool.
// The caller owns the pool and closes it; Client has no Close method.
func Open(ctx context.Context, options ...Option) (*Client, error) {
	// Set default options
	client := &Client{
		logger:      zap.NewNop(),
		pingTimeout: DefaultPingTimeoutSeconds * time.Second,
	}

	// Apply options
	for _, option := range options {
		err := option(client)
		if err != nil {
			return nil, err
		}
	}

	// Validate
	err := client.validate()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (client *Client) Ping(ctx context.Context) error {
	if client.pingTimeout <= 0 {
		return client.db.Ping(ctx)
	}

	ctx, cancelFunc := context.WithTimeout(ctx, client.pingTimeout)
	defer cancelFunc()
	return client.db.Ping(ctx)
}

func (client *Client) validate() error {
	if client.db == nil {
		return fmt.Errorf("db is nil")
	}

	if client.logger == nil {
		return fmt.Errorf("logger is nil")
	}

	return nil
}
