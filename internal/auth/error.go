package auth

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrEmptyContextKey              = errors.New("context key cannot be empty")
	ErrEmptyLogger                  = errors.New("logger cannot be empty")
	ErrEmptySubject                 = errors.New("subject cannot be nil")
	ErrMissingSubjectContext        = errors.New("missing Subject context")
	ErrMissingKeyFromSubjectContext = errors.New("missing key from Subject context")
	ErrForbidden                    = errors.New("forbidden")
	ErrTokenExpired                 = errors.New("token has expired")
	ErrInvalidToken                 = errors.New("invalid token")
	ErrInvalidClaims                = errors.New("invalid token claims")
	ErrRateLimitExceeded            = errors.New("rate limit exceeded")
	ErrAccountLocked                = errors.New("account is locked due to too many failed attempts")
	ErrInsufficientPermissions      = errors.New("insufficient permissions")
	ErrInvalidRole                  = errors.New("invalid or missing role")
)

type Error struct {
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("auth error: %s", e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func HasError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

func GetError(err error) (*Error, error) {
	var e *Error
	if errors.As(err, &e) {
		return e, nil
	}

	return nil, fmt.Errorf("error is not an auth error: %w", err)
}
