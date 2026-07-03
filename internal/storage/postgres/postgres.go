// Package postgres provides a PostgreSQL implementation of the storage interfaces.
// It uses pgx/v5 as the underlying database driver and supports connection pooling,
// transactions, migrations, and comprehensive logging.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

const (
	// DefaultPingTimeoutSeconds is the default timeout for database ping operations
	DefaultPingTimeoutSeconds = 15
)

// dbconn interface defines the specific functions we use from the pxg library
// allows us to bypass a specific type and use *pgx.Conn, *pgxpool.Pool, and a mocking struct
// https://github.com/jackc/pgx/issues/616#issuecomment-535749087
type dbconn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	Ping(ctx context.Context) error
}

// Client implements the storage interfaces for PostgreSQL databases.
// It manages database connections, transactions, and provides logging capabilities.
// Use Open() to create and initialize a new Client instance.
type Client struct {
	logger *zap.Logger
	db     dbconn
	// pool is the underlying pgx connection pool when the client was built
	// with WithPool or WithURI. It is nil for connection- or mock-backed
	// clients. Migrate needs a real pool because goose runs through a
	// database/sql handle wrapped over it (see migration.go).
	pool        *pgxpool.Pool
	closeFunc   func(ctx context.Context) error
	pingTimeout time.Duration
}

// Open creates and initializes a new PostgreSQL database client.
// It accepts a context for timeout control and various configuration options.
// The client must be configured with a database connection using WithURI, WithPool, etc.
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

func (client *Client) Name() string {
	return "postgresql"
}

func (client *Client) Ping(ctx context.Context) error {
	ctx, cancelFunc := context.WithTimeout(ctx, client.pingTimeout)
	defer cancelFunc()
	return client.db.Ping(ctx)
}

func (client *Client) Close() error {
	return client.closeFunc(context.Background())
}

func (client *Client) Logger() *zap.Logger {
	return client.logger
}

func (client *Client) validate() error {
	if client.db == nil {
		return fmt.Errorf("db is nil")
	}

	if client.closeFunc == nil {
		return fmt.Errorf("closeFunc is nil")
	}

	if client.logger == nil {
		return fmt.Errorf("logger is nil")
	}

	return nil
}
