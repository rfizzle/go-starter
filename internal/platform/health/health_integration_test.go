//go:build integration

package health_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap/zaptest"

	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
	"github.com/rfizzle/go-starter/internal/platform/health"
)

func TestAggregatorTracksRealPostgresAvailability(t *testing.T) {
	databaseURI := postgrestest.New(t).EmptyDatabase(t).URI

	newPool := func() *pgxpool.Pool {
		t.Helper()
		pool, err := pgxpool.New(t.Context(), databaseURI)
		if err != nil {
			t.Fatalf("pgxpool.New() error: %v", err)
		}
		return pool
	}

	var activePool atomic.Pointer[pgxpool.Pool]
	initialPool := newPool()
	activePool.Store(initialPool)
	t.Cleanup(func() {
		if pool := activePool.Load(); pool != nil {
			pool.Close()
		}
	})

	aggregator, err := health.New(2*time.Second, zaptest.NewLogger(t), health.Check{
		Name: "postgres",
		Run: func(ctx context.Context) error {
			pool := activePool.Load()
			if pool == nil {
				return errors.New("postgres pool is unavailable")
			}
			return pool.Ping(ctx)
		},
	})
	if err != nil {
		t.Fatalf("health.New() error = %v", err)
	}

	if err := aggregator.Check(t.Context()); err != nil {
		t.Fatalf("healthy PostgreSQL Check() error = %v", err)
	}

	initialPool.Close()
	if aggregator.Ready(t.Context()) {
		t.Fatal("Ready() = true after PostgreSQL pool closed")
	}

	recoveredPool := newPool()
	activePool.Store(recoveredPool)
	if !aggregator.Ready(t.Context()) {
		t.Fatal("Ready() = false after PostgreSQL pool recovery")
	}
}
