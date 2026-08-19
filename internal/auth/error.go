package auth

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrEmptySubject                 = errors.New("subject cannot be nil")
	ErrMissingSubjectContext        = errors.New("missing Subject context")
	ErrMissingKeyFromSubjectContext = errors.New("missing key from Subject context")
	ErrTokenExpired                 = errors.New("token has expired")
	ErrInvalidToken                 = errors.New("invalid token")
	ErrInvalidClaims                = errors.New("invalid token claims")
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
