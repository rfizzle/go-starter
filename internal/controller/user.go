package controller

import (
	"context"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
)

// GetCurrentUser returns the subject encoded in the caller's bearer token. The
// subject was placed on the context by HandleBearerAuth during security
// validation, so reaching this handler already implies a valid token.
func (c *Controller) GetCurrentUser(ctx context.Context) (api.GetCurrentUserRes, error) {
	subject, err := auth.GetSubjectFromContext(ctx, subjectContextKey)
	if err != nil || subject == "" {
		return &api.GetCurrentUserUnauthorized{}, nil
	}

	return &api.CurrentUserResponse{Subject: subject}, nil
}
