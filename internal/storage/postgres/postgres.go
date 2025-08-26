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

// Provider implements the storage interfaces for PostgreSQL databases.
// It manages database connections, transactions, and provides logging capabilities.
// Use Open() to create and initialize a new Provider instance.
type Provider struct {
	logger      *zap.Logger
	db          dbconn
	closeFunc   func(ctx context.Context) error
	pingTimeout time.Duration
}

// Open creates and initializes a new PostgreSQL database provider.
// It accepts a context for timeout control and various configuration options.
// The provider must be configured with a database connection using WithURI, WithPool, etc.
func Open(ctx context.Context, options ...Option) (*Provider, error) {
	// Set default options
	provider := &Provider{
		logger:      zap.NewNop(),
		pingTimeout: DefaultPingTimeoutSeconds * time.Second,
	}

	// Apply options
	for _, option := range options {
		err := option(provider)
		if err != nil {
			return nil, err
		}
	}

	// Validate
	err := provider.validate()
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func (provider *Provider) Name() string {
	return "postgresql"
}

func (provider *Provider) Ping(ctx context.Context) error {
	ctx, cancelFunc := context.WithTimeout(ctx, provider.pingTimeout)
	defer cancelFunc()
	return provider.db.Ping(ctx)
}

func (provider *Provider) Close() error {
	return provider.closeFunc(context.Background())
}

func (provider *Provider) Logger() *zap.Logger {
	return provider.logger
}

func (provider *Provider) validate() error {
	if provider.db == nil {
		return fmt.Errorf("db is nil")
	}

	if provider.closeFunc == nil {
		return fmt.Errorf("closeFunc is nil")
	}

	if provider.logger == nil {
		return fmt.Errorf("logger is nil")
	}

	return nil
}
