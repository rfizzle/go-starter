package app

import (
	"context"
	"embed"
	"errors"
	"sync/atomic"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/domain"
	"github.com/rfizzle/go-starter/internal/modules/credential"
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

type stubCredentialReader struct {
	credential domain.Credential
	err        error
}

func (s *stubCredentialReader) CredentialByEmail(context.Context, string) (domain.Credential, error) {
	return s.credential, s.err
}

func TestCredentialAuthenticator_Authenticate(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	readerErr := errors.New("database unavailable")

	tests := []struct {
		name        string
		reader      *stubCredentialReader
		password    string
		wantSubject string
		wantErr     error
	}{
		{
			name:     "unknown email maps to controller invalid credentials",
			reader:   &stubCredentialReader{err: domain.ErrNotFound},
			password: "correct-password",
			wantErr:  controller.ErrInvalidCredentials,
		},
		{
			name: "wrong password maps to controller invalid credentials",
			reader: &stubCredentialReader{credential: domain.Credential{
				UserID:       "user-1",
				Email:        "ada@example.com",
				PasswordHash: string(hash),
			}},
			password: "wrong-password",
			wantErr:  controller.ErrInvalidCredentials,
		},
		{
			name:     "reader error remains distinct",
			reader:   &stubCredentialReader{err: readerErr},
			password: "correct-password",
			wantErr:  readerErr,
		},
		{
			name: "success returns credential email",
			reader: &stubCredentialReader{credential: domain.Credential{
				UserID:       "user-1",
				Email:        "ada@example.com",
				PasswordHash: string(hash),
			}},
			password:    "correct-password",
			wantSubject: "ada@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := credential.NewService(tt.reader)
			if err != nil {
				t.Fatalf("credential.NewService() error: %v", err)
			}
			authenticator := &credentialAuthenticator{service: service}

			subject, err := authenticator.Authenticate(t.Context(), "ada@example.com", tt.password)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Authenticate() error = %v, want error matching %v", err, tt.wantErr)
				}
				if errors.Is(err, controller.ErrInvalidCredentials) != errors.Is(tt.wantErr, controller.ErrInvalidCredentials) {
					t.Fatalf("Authenticate() invalid credential mapping = %t, want %t", errors.Is(err, controller.ErrInvalidCredentials), errors.Is(tt.wantErr, controller.ErrInvalidCredentials))
				}
				return
			}
			if err != nil {
				t.Fatalf("Authenticate() error: %v", err)
			}
			if subject != tt.wantSubject {
				t.Errorf("Authenticate() subject = %q, want %q", subject, tt.wantSubject)
			}
		})
	}
}
