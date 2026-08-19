package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/domain"
	"github.com/rfizzle/go-starter/internal/modules/credential"
)

func TestMigrateOnStartup_DisabledSkipsMigration(t *testing.T) {
	calls := 0
	migrate := func(context.Context) error {
		calls++
		return nil
	}

	if err := migrateOnStartup(t.Context(), false, migrate); err != nil {
		t.Fatalf("migrateOnStartup() error = %v, want nil", err)
	}
	if calls != 0 {
		t.Errorf("migrate calls = %d, want 0", calls)
	}
}

func TestMigrateOnStartup_EnabledPropagatesFailure(t *testing.T) {
	migrationErr := errors.New("migration failed")
	calls := 0
	migrate := func(context.Context) error {
		calls++
		return migrationErr
	}

	err := migrateOnStartup(t.Context(), true, migrate)
	if !errors.Is(err, migrationErr) {
		t.Fatalf("migrateOnStartup() error = %v, want error wrapping %v", err, migrationErr)
	}
	if calls != 1 {
		t.Errorf("migrate calls = %d, want 1", calls)
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

type stubCredentialReader struct {
	err error
}

func (s *stubCredentialReader) CredentialByEmail(context.Context, string) (domain.Credential, error) {
	return domain.Credential{}, s.err
}

func TestCredentialAuthenticator_Authenticate(t *testing.T) {
	// The adapter's only behaviour of its own is translating the credential
	// module's sentinel into the controller's; every other path is covered by
	// internal/modules/credential/service_test.go.
	service, err := credential.NewService(&stubCredentialReader{err: domain.ErrNotFound}, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("credential.NewService() error: %v", err)
	}
	authenticator := &credentialAuthenticator{service: service}

	_, err = authenticator.Authenticate(t.Context(), "ada@example.com", "any-password")

	if !errors.Is(err, controller.ErrInvalidCredentials) {
		t.Fatalf("Authenticate() error = %v, want controller.ErrInvalidCredentials", err)
	}
	if errors.Is(err, credential.ErrInvalidCredentials) {
		t.Fatalf("Authenticate() error = %v, credential.ErrInvalidCredentials must not leak past the adapter", err)
	}
}
