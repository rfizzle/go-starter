package postgres

import (
	"context"

	"github.com/rfizzle/go-starter/internal/storage"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
)

func (provider *Provider) Query(ctx context.Context, sql string, args ...interface{}) (storage.RowScanner, error) {
	rows, err := provider._query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return &internalScanner{rows: rows}, nil
}

func (provider *Provider) Execute(ctx context.Context, sql string, args ...interface{}) error {
	return provider._exec(ctx, sql, args...)
}

func (provider *Provider) _query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	provider.logger.Debug("database query", zap.String("sql", sql), zap.Any("args", args))
	rows, err := provider.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (provider *Provider) _exec(ctx context.Context, sql string, args ...interface{}) error {
	provider.logger.Debug("database exec", zap.String("sql", sql), zap.Any("args", args))
	_, err := provider.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}
