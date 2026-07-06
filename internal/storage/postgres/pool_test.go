package postgres

import (
	"context"
	"math"
	"testing"
	"time"

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

func TestClampInt32(t *testing.T) {
	assert.Equal(t, int32(10), clampInt32(10))
	assert.Equal(t, int32(math.MaxInt32), clampInt32(math.MaxInt32))
	assert.Equal(t, int32(math.MaxInt32), clampInt32(math.MaxInt32+1))
}
