package auth

import (
	"crypto/subtle"
	"time"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// TokenTimeout specifies how long tokens are valid
	TokenTimeout time.Duration
	// MaxLoginAttempts specifies max failed login attempts before lockout
	MaxLoginAttempts int
	// LockoutDuration specifies how long accounts are locked
	LockoutDuration time.Duration
	// RequireHTTPS enforces HTTPS-only cookies
	RequireHTTPS bool
	// SameSitePolicy for cookies
	SameSitePolicy string
}

// DefaultSecurityConfig returns sensible security defaults
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		TokenTimeout:     24 * time.Hour,
		MaxLoginAttempts: 5,
		LockoutDuration:  15 * time.Minute,
		RequireHTTPS:     true,
		SameSitePolicy:   "Strict",
	}
}

// SecureCompare performs constant-time string comparison to prevent timing attacks
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// TokenClaims represents standard JWT claims with additional fields
type TokenClaims struct {
	Subject   string    `json:"sub"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	Roles     []string  `json:"roles,omitempty"`
	Scopes    []string  `json:"scopes,omitempty"`
}

// IsExpired checks if the token has expired
func (c TokenClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// HasRole checks if the token contains a specific role
func (c TokenClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasScope checks if the token contains a specific scope
func (c TokenClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}
