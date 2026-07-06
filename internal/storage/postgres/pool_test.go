package postgres

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid URI returns error", func(t *testing.T) {
		pool, err := NewPool(ctx, "invalid://uri", PoolConfig{})
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("applies tuning knobs to the parsed config", func(t *testing.T) {
		// NewWithConfig connects lazily, so a valid URI to an unreachable host
		// still returns a usable pool whose config we can inspect.
		pool, err := NewPool(ctx, "postgres://user:pass@localhost:5432/db", PoolConfig{
			MaxConns:                 25,
			MinConns:                 5,
			MaxConnLifetimeSeconds:   1800,
			MaxConnIdleTimeSeconds:   300,
			HealthCheckPeriodSeconds: 30,
		})
		require.NoError(t, err)
		require.NotNil(t, pool)
		defer pool.Close()

		cfg := pool.Config()
		assert.Equal(t, int32(25), cfg.MaxConns)
		assert.Equal(t, int32(5), cfg.MinConns)
		assert.Equal(t, 1800*time.Second, cfg.MaxConnLifetime)
		assert.Equal(t, 300*time.Second, cfg.MaxConnIdleTime)
		assert.Equal(t, 30*time.Second, cfg.HealthCheckPeriod)
	})

	t.Run("min-conns exceeding max-conns returns error", func(t *testing.T) {
		pool, err := NewPool(ctx, "postgres://user:pass@localhost:5432/db", PoolConfig{
			MaxConns: 10,
			MinConns: 50,
		})
		require.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "min-conns")
	})

	t.Run("zero values leave pgx defaults in place", func(t *testing.T) {
		pool, err := NewPool(ctx, "postgres://user:pass@localhost:5432/db", PoolConfig{})
		require.NoError(t, err)
		require.NotNil(t, pool)
		defer pool.Close()

		cfg := pool.Config()
		// pgx defaults: MaxConns is max(4, NumCPU), the rest are fixed.
		assert.GreaterOrEqual(t, cfg.MaxConns, int32(4))
		assert.Equal(t, int32(0), cfg.MinConns)
		assert.Equal(t, time.Hour, cfg.MaxConnLifetime)
		assert.Equal(t, 30*time.Minute, cfg.MaxConnIdleTime)
		assert.Equal(t, time.Minute, cfg.HealthCheckPeriod)
	})
}

func TestClosePool(t *testing.T) {
	ctx := context.Background()

	// A lazily-connected pool to an unreachable host holds no live connections,
	// so Close returns promptly and ClosePool should not report a timeout.
	newIdlePool := func(t *testing.T) *pgxpool.Pool {
		t.Helper()
		pool, err := NewPool(ctx, "postgres://user:pass@localhost:5432/db", PoolConfig{})
		require.NoError(t, err)
		require.NotNil(t, pool)
		return pool
	}

	t.Run("returns nil when close completes within the timeout", func(t *testing.T) {
		pool := newIdlePool(t)
		assert.NoError(t, ClosePool(pool, 5*time.Second))
	})

	t.Run("non-positive timeout waits for close and returns nil", func(t *testing.T) {
		pool := newIdlePool(t)
		assert.NoError(t, ClosePool(pool, 0))
	})

	t.Run("double close is safe", func(t *testing.T) {
		// pgxpool.Close guards with sync.Once, so a second ClosePool after the
		// pool is already closed must still return promptly without panicking.
		pool := newIdlePool(t)
		require.NoError(t, ClosePool(pool, 5*time.Second))
		assert.NoError(t, ClosePool(pool, 5*time.Second))
	})
}

func TestClampInt32(t *testing.T) {
	assert.Equal(t, int32(10), clampInt32(10))
	assert.Equal(t, int32(math.MaxInt32), clampInt32(math.MaxInt32))
	assert.Equal(t, int32(math.MaxInt32), clampInt32(math.MaxInt32+1))
}
