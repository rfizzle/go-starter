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

type Option func(p *Provider) error

func WithURI(uri string) Option {
	return func(p *Provider) error {
		conn, err := pgxpool.New(p.ctx, uri)
		if err != nil {
			return err
		}

		p.db = conn
		p.closeFunc = func(ctx context.Context) error {
			conn.Close()
			return nil
		}

		return nil
	}
}

func WithConnection(conn *pgx.Conn) Option {
	return func(p *Provider) error {
		p.db = conn
		p.closeFunc = func(ctx context.Context) error {
			return conn.Close(ctx)
		}

		return nil
	}
}

func WithPool(pool *pgxpool.Pool) Option {
	return func(p *Provider) error {
		p.db = pool
		p.closeFunc = func(ctx context.Context) error {
			pool.Close()
			return nil
		}

		return nil
	}
}

func WithMockPool(pool pgxmock.PgxPoolIface) Option {
	return func(p *Provider) error {
		p.db = pool
		p.closeFunc = func(ctx context.Context) error {
			return nil
		}

		return nil
	}
}

func WithPingTimeout(timeout int) Option {
	return func(p *Provider) error {
		p.pingTimeout = time.Duration(timeout) * time.Second
		return nil
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(p *Provider) error {
		if logger == nil {
			return fmt.Errorf("logger cannot be nil")
		}

		p.logger = logger
		return nil
	}
}
