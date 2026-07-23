package app

import (
	"context"
	"embed"
	"errors"
	"sync/atomic"
	"testing"
)

type fakeMigrator struct {
	err   error
	calls int
}

func (m *fakeMigrator) Migrate(context.Context, embed.FS) error {
	m.calls++
	return m.err
}

func TestMigrateOnStartup_DisabledSkipsMigration(t *testing.T) {
	store := &fakeMigrator{}

	if err := migrateOnStartup(t.Context(), false, store); err != nil {
		t.Fatalf("migrateOnStartup() error = %v, want nil", err)
	}
	if store.calls != 0 {
		t.Errorf("Migrate() calls = %d, want 0", store.calls)
	}
}

func TestMigrateOnStartup_EnabledPropagatesFailure(t *testing.T) {
	migrationErr := errors.New("migration failed")
	store := &fakeMigrator{err: migrationErr}

	err := migrateOnStartup(t.Context(), true, store)
	if !errors.Is(err, migrationErr) {
		t.Fatalf("migrateOnStartup() error = %v, want error wrapping %v", err, migrationErr)
	}
	if store.calls != 1 {
		t.Errorf("Migrate() calls = %d, want 1", store.calls)
	}
}

type fakeReadiness struct {
	ready bool
	calls int
}

func (readiness *fakeReadiness) Ready(context.Context) bool {
	readiness.calls++
	return readiness.ready
}

func TestAppReadinessTracksLifecycleAndDependencyRecovery(t *testing.T) {
	lifecycle := &atomic.Bool{}
	dependencies := &fakeReadiness{ready: true}
	readiness := &appReadiness{
		lifecycle:    lifecycle,
		dependencies: dependencies,
	}

	if readiness.Ready(t.Context()) {
		t.Fatal("Ready() = true during startup")
	}
	if dependencies.calls != 0 {
		t.Fatalf("dependency calls during startup = %d, want 0", dependencies.calls)
	}

	lifecycle.Store(true)
	if !readiness.Ready(t.Context()) {
		t.Fatal("Ready() = false after healthy startup")
	}

	dependencies.ready = false
	if readiness.Ready(t.Context()) {
		t.Fatal("Ready() = true while dependency is unavailable")
	}

	dependencies.ready = true
	if !readiness.Ready(t.Context()) {
		t.Fatal("Ready() = false after dependency recovery")
	}

	lifecycle.Store(false)
	callsBeforeShutdown := dependencies.calls
	if readiness.Ready(t.Context()) {
		t.Fatal("Ready() = true during shutdown")
	}
	if dependencies.calls != callsBeforeShutdown {
		t.Errorf(
			"dependency calls during shutdown = %d, want %d",
			dependencies.calls,
			callsBeforeShutdown,
		)
	}
}
