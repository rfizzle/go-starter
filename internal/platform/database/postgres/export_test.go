package postgres

import (
	"context"

	"github.com/pashagolub/pgxmock/v4"
)

// WithMockPool configures the client to use a mock PostgreSQL pool for testing.
// It lives in a _test.go file so the pgxmock dependency stays out of the
// production build.
func WithMockPool(pool pgxmock.PgxPoolIface) Option {
	return func(p *Client) error {
		p.db = pool
		p.closeFunc = func(ctx context.Context) error {
			return nil
		}

		return nil
	}
}
