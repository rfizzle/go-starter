package controller

import (
	"context"
	"errors"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/rfizzle/go-starter/internal/domain"
)

// GetCurrentUser returns the subject encoded in the caller's bearer token,
// enriched with the matching user record when one exists. The subject was
// placed on the context by HandleBearerAuth during security validation, so
// reaching this handler already implies a valid token.
//
// The subject is treated as the user's email for the lookup. When no user
// reader is configured, or no user matches, the endpoint still succeeds with
// just the subject — the starter is usable before any users are seeded.
func (c *Controller) GetCurrentUser(ctx context.Context) (api.GetCurrentUserRes, error) {
	subject, err := auth.GetSubjectFromContext(ctx, subjectContextKey)
	if err != nil || subject == "" {
		return &api.GetCurrentUserUnauthorized{}, nil
	}

	resp := &api.CurrentUserResponse{Subject: subject}

	if c.options.Users != nil {
		u, lookupErr := c.options.Users.ByEmail(ctx, subject)
		switch {
		case lookupErr == nil:
			resp.Email = api.NewOptString(u.Email)
			resp.FirstName = api.NewOptString(u.FirstName)
			resp.LastName = api.NewOptString(u.LastName)
		case errors.Is(lookupErr, domain.ErrNotFound):
			// No stored user for this subject yet; return the subject alone.
		default:
			return nil, lookupErr
		}
	}

	return resp, nil
}
