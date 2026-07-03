package postgres

import (
	"context"

	"github.com/rfizzle/go-starter/internal/storage"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
)

func (client *Client) Query(ctx context.Context, sql string, args ...interface{}) (storage.RowScanner, error) {
	rows, err := client._query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return &internalScanner{rows: rows}, nil
}

func (client *Client) Execute(ctx context.Context, sql string, args ...interface{}) error {
	return client._exec(ctx, sql, args...)
}

func (client *Client) _query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	client.logger.Debug("database query", zap.String("sql", sql), zap.Any("args", args))
	rows, err := client.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (client *Client) _exec(ctx context.Context, sql string, args ...interface{}) error {
	client.logger.Debug("database exec", zap.String("sql", sql), zap.Any("args", args))
	_, err := client.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}
