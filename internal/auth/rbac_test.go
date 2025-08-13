package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Mock role provider for testing
type mockRoleProvider struct {
	roles map[string][]string
	err   error
}

func (m *mockRoleProvider) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	roles, exists := m.roles[userID]
	if !exists {
		return []string{}, nil
	}
	return roles, nil
}

// Mock authorization provider for testing
type mockAuthorizationProvider struct {
	permissions map[string]map[Permission]bool
	err         error
}

func (m *mockAuthorizationProvider) HasPermission(ctx context.Context, userID string, permission Permission) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	userPerms, exists := m.permissions[userID]
	if !exists {
		return false, nil
	}
	return userPerms[permission], nil
}

func TestRequireRole(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	roleProvider := &mockRoleProvider{
		roles: map[string][]string{
			"user1": {"admin", "user"},
			"user2": {"user"},
			"user3": {},
		},
	}

	middleware := auth.RequireRole(roleProvider, "admin")

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "user with admin role",
			userID:         "user1",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "user without admin role",
			userID:         "user2",
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
		{
			name:           "user with no roles",
			userID:         "user3",
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with user
			ctx, err := auth.newSubjectContext(context.Background(), tt.userID)
			assert.NoError(t, err)

			req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.shouldCallNext, nextCalled)
		})
	}
}

func TestRequireRole_NoAuthenticatedUser(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	roleProvider := &mockRoleProvider{}
	middleware := auth.RequireRole(roleProvider, "admin")

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextCalled)
}

func TestRequireRole_RoleProviderError(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	roleProvider := &mockRoleProvider{
		err: errors.New("database error"),
	}
	middleware := auth.RequireRole(roleProvider, "admin")

	ctx, err := auth.newSubjectContext(context.Background(), "user1")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextCalled)
}

func TestRequirePermission(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	permission := Permission{Resource: "users", Action: "read"}
	authzProvider := &mockAuthorizationProvider{
		permissions: map[string]map[Permission]bool{
			"user1": {permission: true},
			"user2": {permission: false},
		},
	}

	middleware := auth.RequirePermission(authzProvider, permission)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "user with permission",
			userID:         "user1",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "user without permission",
			userID:         "user2",
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := auth.newSubjectContext(context.Background(), tt.userID)
			assert.NoError(t, err)

			req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.shouldCallNext, nextCalled)
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	roleProvider := &mockRoleProvider{
		roles: map[string][]string{
			"user1": {"admin"},
			"user2": {"moderator"},
			"user3": {"user"},
			"user4": {"guest"},
		},
	}

	middleware := auth.RequireAnyRole(roleProvider, "admin", "moderator")

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "user with admin role",
			userID:         "user1",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "user with moderator role",
			userID:         "user2",
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "user with neither role",
			userID:         "user3",
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
		{
			name:           "user with different role",
			userID:         "user4",
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := auth.newSubjectContext(context.Background(), tt.userID)
			assert.NoError(t, err)

			req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.shouldCallNext, nextCalled)
		})
	}
}

func TestRequirePermission_AuthorizationProviderError(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	authzProvider := &mockAuthorizationProvider{
		err: errors.New("authorization service error"),
	}
	permission := Permission{Resource: "users", Action: "read"}
	middleware := auth.RequirePermission(authzProvider, permission)

	ctx, err := auth.newSubjectContext(context.Background(), "user1")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := middleware(nextHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextCalled)
}
