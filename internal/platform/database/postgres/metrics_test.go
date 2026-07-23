package postgres

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoolCollector(t *testing.T) {
	ctx := context.Background()

	// A lazily-connected pool to an unreachable host holds no live connections,
	// so its Stat() snapshot is well-defined (all zero-ish) without a database.
	pool, err := NewPool(ctx, "postgres://user:pass@localhost:5432/db", PoolConfig{MaxConns: 10})
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer pool.Close()

	collector := NewPoolCollector(pool)

	// Every Desc emitted by Describe must be produced by Collect and vice versa,
	// so derive the expected family count from Describe rather than hardcoding a
	// magic number that silently drifts when a metric is added or removed.
	describeCount := func() int {
		ch := make(chan *prometheus.Desc, 32)
		collector.Describe(ch)
		close(ch)

		count := 0
		for range ch {
			count++
		}
		return count
	}

	t.Run("describe emits at least one desc", func(t *testing.T) {
		assert.Positive(t, describeCount())
	})

	t.Run("registers cleanly and gathers a family per described metric", func(t *testing.T) {
		// Gathering through a dedicated registry exercises registration and the
		// registry's metric-consistency checks (duplicate names, inconsistent
		// label sets) without touching the process-wide default registry.
		reg := prometheus.NewRegistry()
		require.NoError(t, reg.Register(collector))

		families, err := reg.Gather()
		require.NoError(t, err)
		assert.Len(t, families, describeCount())

		// max_connections reflects the configured cap, confirming the collector
		// reads live values off pool.Stat() rather than emitting placeholders.
		values := map[string]float64{}
		for _, mf := range families {
			require.Len(t, mf.GetMetric(), 1)
			m := mf.GetMetric()[0]
			switch {
			case m.GetGauge() != nil:
				values[mf.GetName()] = m.GetGauge().GetValue()
			case m.GetCounter() != nil:
				values[mf.GetName()] = m.GetCounter().GetValue()
			}
		}
		assert.Equal(t, float64(10), values["db_pool_max_connections"])
		assert.Equal(t, float64(0), values["db_pool_acquired_connections"])
	})
}
