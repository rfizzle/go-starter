package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	defaultUID = "abc12345678"
)

func TestNewSubjectContext(t *testing.T) {
	ctx := context.Background()
	authService, err := NewService(context.Background(), &mockValidator{}, Logger(zap.NewNop()))
	assert.NotNil(t, authService)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	svc := authService.(*service)
	newCtx, err := newAuthSubjectContext(ctx, svc.contextKey, defaultUID)
	assert.NotNil(t, newCtx)
	assert.Nil(t, err)
	assert.NotEqual(t, ctx, newCtx, "context should be different")
}

func TestNewSubjectContext_InvalidType(t *testing.T) {
	ctx := context.Background()
	authService, err := NewService(context.Background(), &mockValidator{}, Logger(zap.NewNop()))
	assert.NotNil(t, authService)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	svc := authService.(*service)
	newCtx, expectedErr := newAuthSubjectContext(ctx, svc.contextKey, "")
	assert.Nil(t, newCtx)
	assert.NotNil(t, expectedErr)
	authErr, err := GetError(expectedErr)
	assert.NotNil(t, authErr)
	assert.Nil(t, err)
	assert.Equal(t, authErr.Err, ErrEmptySubject)
}

func TestNewSubjectContextWithCustomKey(t *testing.T) {
	ctx := context.Background()
	authService, err := NewService(context.Background(), &mockValidator{}, Logger(zap.NewNop()), ContextKey("custom"))
	assert.NotNil(t, authService)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	newCtx, err := newAuthSubjectContext(ctx, "custom", defaultUID)
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