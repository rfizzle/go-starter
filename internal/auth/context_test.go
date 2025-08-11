package auth

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

const (
	defaultUID = "abc12345678"
)

func TestNewSubjectContext(t *testing.T) {
	ctx := context.Background()
	a, err := NewAuth(&mockValidator{}, WithLogger(zap.NewNop()))
	assert.NotNil(t, a)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, err := a.newSubjectContext(ctx, defaultUID)
	assert.NotNil(t, newCtx)
	assert.Nil(t, err)
	assert.NotEqualf(t, ctx, newCtx, "context should be different")
}

func TestNewSubjectContext_InvalidType(t *testing.T) {
	ctx := context.Background()
	a, err := NewAuth(&mockValidator{}, WithLogger(zap.NewNop()))
	assert.NotNil(t, a)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, expectedErr := a.newSubjectContext(ctx, "")
	assert.Nil(t, newCtx)
	assert.NotNil(t, expectedErr)
	authErr, err := GetError(expectedErr)
	assert.NotNil(t, authErr)
	assert.Nil(t, err)
	assert.Equal(t, authErr.Err, ErrEmptySubject)
}

func TestNewSubjectContextWithCustomKey(t *testing.T) {
	ctx := context.Background()
	a, err := NewAuth(&mockValidator{}, WithLogger(zap.NewNop()), WithContextKey("custom"))
	assert.NotNil(t, a)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, err := a.newSubjectContext(ctx, defaultUID)
	assert.NotNil(t, newCtx)
	assert.Nil(t, err)
	assert.NotEqualf(t, ctx, newCtx, "context should be different")
	subject, err := subjectFromContext(newCtx, "custom")
	assert.NotEmpty(t, subject)
	assert.Nil(t, err)
	assert.Equal(t, defaultUID, subject)
	failSubject, err := subjectFromContext(newCtx, "default")
	assert.Empty(t, failSubject)
	assert.NotNil(t, err)
}
