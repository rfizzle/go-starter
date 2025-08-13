package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"go.uber.org/zap"
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

type Provider struct {
	logger      *zap.Logger
	db          dbconn
	closeFunc   func(ctx context.Context) error
	ctx         context.Context
	pingTimeout time.Duration
}

func Open(ctx context.Context, options ...Option) (*Provider, error) {
	// Set default options
	const (
		defaultPingTimeoutSeconds = 15
	)

	provider := &Provider{
		ctx:         ctx,
		logger:      zap.NewNop(),
		pingTimeout: defaultPingTimeoutSeconds * time.Second,
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

func (provider *Provider) Ping() error {
	ctx, cancelFunc := context.WithTimeout(provider.ctx, provider.pingTimeout)
	defer cancelFunc()
	return provider.db.Ping(ctx)
}

func (provider *Provider) Disconnect() error {
	return provider.closeFunc(provider.ctx)
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
