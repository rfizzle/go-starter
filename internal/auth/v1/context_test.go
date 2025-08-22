package v1

import (
	"context"
	"testing"

	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	defaultUID = "abc12345678"
)

func TestNewSubjectContext(t *testing.T) {
	ctx := context.Background()
	a, err := NewAuth(&mockValidator{}, WithLoggerLegacy(zap.NewNop()))
	assert.NotNil(t, a)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, err := newAuthSubjectContext(ctx, a.contextKey, defaultUID)
	assert.NotNil(t, newCtx)
	assert.Nil(t, err)
	assert.NotEqual(t, ctx, newCtx, "context should be different")
}

func TestNewSubjectContext_InvalidType(t *testing.T) {
	ctx := context.Background()
	a, err := NewAuth(&mockValidator{}, WithLoggerLegacy(zap.NewNop()))
	assert.NotNil(t, a)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, expectedErr := newAuthSubjectContext(ctx, a.contextKey, "")
	assert.Nil(t, newCtx)
	assert.NotNil(t, expectedErr)
	authErr, err := auth.GetError(expectedErr)
	assert.NotNil(t, authErr)
	assert.Nil(t, err)
	assert.Equal(t, authErr.Err, auth.ErrEmptySubject)
}

func TestNewSubjectContextWithCustomKey(t *testing.T) {
	ctx := context.Background()
	a, err := NewAuth(&mockValidator{}, WithLoggerLegacy(zap.NewNop()), WithContextKeyLegacy("custom"))
	assert.NotNil(t, a)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, err := newAuthSubjectContext(ctx, a.contextKey, defaultUID)
	assert.NotNil(t, newCtx)
	assert.Nil(t, err)
	assert.NotEqual(t, ctx, newCtx, "context should be different")
	subject, err := subjectFromContext(newCtx, "custom")
	assert.NotEmpty(t, subject)
	assert.Nil(t, err)
	assert.Equal(t, defaultUID, subject)
	failSubject, err := subjectFromContext(newCtx, "default")
	assert.Empty(t, failSubject)
	assert.NotNil(t, err)
}