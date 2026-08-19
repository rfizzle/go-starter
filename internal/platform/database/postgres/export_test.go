package postgres

import (
	"context"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// MigrateFS exposes the private migration seam to external-package integration
// tests without adding fs.FS to the production API.
func MigrateFS(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger, migrationFS fs.FS) error {
	return migrate(ctx, pool, logger, migrationFS)
}
