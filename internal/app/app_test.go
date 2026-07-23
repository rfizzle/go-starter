package app

import (
	"context"
	"embed"
	"errors"
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
