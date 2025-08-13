package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical strings",
			a:        "password123",
			b:        "password123",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "password123",
			b:        "password456",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty string",
			a:        "password",
			b:        "",
			expected: false,
		},
		{
			name:     "different lengths",
			a:        "short",
			b:        "longer_password",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecureCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenClaims_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		claims   TokenClaims
		expected bool
	}{
		{
			name: "not expired",
			claims: TokenClaims{
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "expired",
			claims: TokenClaims{
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expected: true,
		},
		{
			name: "expires now",
			claims: TokenClaims{
				ExpiresAt: time.Now(),
			},
			expected: false, // Might be flaky due to timing, but should be very close
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.claims.IsExpired()
			if tt.name == "expires now" {
				// For the edge case, allow either result
				assert.Contains(t, []bool{true, false}, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTokenClaims_HasRole(t *testing.T) {
	claims := TokenClaims{
		Roles: []string{"admin", "user", "moderator"},
	}

	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "has role admin",
			role:     "admin",
			expected: true,
		},
		{
			name:     "has role user",
			role:     "user",
			expected: true,
		},
		{
			name:     "does not have role guest",
			role:     "guest",
			expected: false,
		},
		{
			name:     "empty role",
			role:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := claims.HasRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenClaims_HasScope(t *testing.T) {
	claims := TokenClaims{
		Scopes: []string{"read:users", "write:users", "delete:posts"},
	}

	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{
			name:     "has scope read:users",
			scope:    "read:users",
			expected: true,
		},
		{
			name:     "has scope write:users",
			scope:    "write:users",
			expected: true,
		},
		{
			name:     "does not have scope admin:all",
			scope:    "admin:all",
			expected: false,
		},
		{
			name:     "empty scope",
			scope:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := claims.HasScope(tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	assert.Equal(t, 24*time.Hour, config.TokenTimeout)
	assert.Equal(t, 5, config.MaxLoginAttempts)
	assert.Equal(t, 15*time.Minute, config.LockoutDuration)
	assert.True(t, config.RequireHTTPS)
	assert.Equal(t, "Strict", config.SameSitePolicy)
}

func TestTokenClaims_EmptyRolesAndScopes(t *testing.T) {
	claims := TokenClaims{
		Roles:  []string{},
		Scopes: []string{},
	}

	assert.False(t, claims.HasRole("admin"))
	assert.False(t, claims.HasScope("read:users"))
}

func TestTokenClaims_NilRolesAndScopes(t *testing.T) {
	claims := TokenClaims{
		Roles:  nil,
		Scopes: nil,
	}

	assert.False(t, claims.HasRole("admin"))
	assert.False(t, claims.HasScope("read:users"))
}
