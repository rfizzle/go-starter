package postgres

import (
	"github.com/rfizzle/go-starter/internal/storage"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
)

func (provider *Provider) Query(sql string, args ...interface{}) (storage.Scanner, error) {
	rows, err := provider._query(sql, args...)
	if err != nil {
		return nil, err
	}

	return &internalScanner{rows: rows}, nil
}

func (provider *Provider) Execute(sql string, args ...interface{}) error {
	return provider._exec(sql, args...)
}

func (provider *Provider) _query(sql string, args ...interface{}) (pgx.Rows, error) {
	provider.logger.Debug("database query", zap.String("sql", sql), zap.Any("args", args))
	rows, err := provider.db.Query(provider.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (provider *Provider) _exec(sql string, args ...interface{}) error {
	provider.logger.Debug("database exec", zap.String("sql", sql), zap.Any("args", args))
	_, err := provider.db.Exec(provider.ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}
