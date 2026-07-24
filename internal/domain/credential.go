// Package domain defines business data, shared repository contracts, and
// business-facing errors. It is an infrastructure-free leaf package.
package domain

// Credential is the minimal authentication-relevant projection of a user. It
// exists only so authentication callers can read a password hash without
// widening domain.User or domain.UserRepository, which must never carry one.
type Credential struct {
	UserID       string
	Email        string
	PasswordHash string
}
