package postgres

import (
	"context"
	"io/fs"

	"github.com/pashagolub/pgxmock/v4"
)

// WithMockPool configures the client to use a mock PostgreSQL pool for testing.
// It lives in a _test.go file so the pgxmock dependency stays out of the
// production build.
func WithMockPool(pool pgxmock.PgxPoolIface) Option {
	return func(p *Client) error {
		p.db = pool
		return nil
	}
}

// MigrateFS exposes the private migration seam to external-package integration
// tests without adding fs.FS to the production API.
func MigrateFS(client *Client, ctx context.Context, migrationFS fs.FS) error {
	return client.migrate(ctx, migrationFS)
}
