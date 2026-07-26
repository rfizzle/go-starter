package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultUID = "abc12345678"
)

// The subject carrier is exercised through WithSubject/GetSubjectFromContext,
// the pair the generated-transport edge actually calls.
func TestWithSubjectRoundTrip(t *testing.T) {
	ctx := context.Background()

	newCtx, err := WithSubject(ctx, "custom", defaultUID)
	assert.NoError(t, err)
	assert.NotEqual(t, ctx, newCtx, "context should be different")

	subject, err := GetSubjectFromContext(newCtx, "custom")
	assert.NoError(t, err)
	assert.Equal(t, defaultUID, subject)
}

func TestWithSubjectRejectsEmptySubject(t *testing.T) {
	newCtx, err := WithSubject(context.Background(), "custom", "")
	assert.Nil(t, newCtx)

	authErr, getErr := GetError(err)
	assert.NoError(t, getErr)
	assert.Equal(t, ErrEmptySubject, authErr.Err)
}

func TestGetSubjectFromContextMissing(t *testing.T) {
	ctx, err := WithSubject(context.Background(), "custom", defaultUID)
	assert.NoError(t, err)

	// A key that was never written must not fall back to another key's subject.
	subject, err := GetSubjectFromContext(ctx, "default")
	assert.Empty(t, subject)
	assert.ErrorIs(t, err, ErrMissingKeyFromSubjectContext)

	// A context that carries no subject at all is distinguishable.
	subject, err = GetSubjectFromContext(context.Background(), "custom")
	assert.Empty(t, subject)
	assert.ErrorIs(t, err, ErrMissingSubjectContext)
}

// A derived context must not mutate a map still visible through its parent.
func TestWithSubjectDoesNotMutateParent(t *testing.T) {
	parent, err := WithSubject(context.Background(), "first", defaultUID)
	assert.NoError(t, err)

	child, err := WithSubject(parent, "second", "other")
	assert.NoError(t, err)

	subject, err := GetSubjectFromContext(child, "first")
	assert.NoError(t, err)
	assert.Equal(t, defaultUID, subject)

	_, err = GetSubjectFromContext(parent, "second")
	assert.ErrorIs(t, err, ErrMissingKeyFromSubjectContext)
}
