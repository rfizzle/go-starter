package auth

import (
	"net/http"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/assert"
)

type mockGenerator struct {
	Token  string
	Cookie *http.Cookie
}

func (mG *mockGenerator) NewToken(subject string) (string, error) {
	return mG.Token, nil
}

func (mG *mockGenerator) NewCookie(subject string) (*http.Cookie, error) {
	return mG.Cookie, nil
}

type mockValidator struct {
	Subject string
	Err     error
}

func (mV *mockValidator) Validate(r *http.Request) (string, error) {
	return mV.Subject, nil
}

func TestMockGenerator_NewToken(t *testing.T) {
	expectedToken := "mockToken"
	generator := &mockGenerator{Token: expectedToken}

	token, err := generator.NewToken("subject")
	assert.Nil(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestMockGenerator_NewCookie(t *testing.T) {
	expectedCookie := &http.Cookie{Name: "mockCookie", Value: "mockValue"}
	generator := &mockGenerator{Cookie: expectedCookie}

	cookie, err := generator.NewCookie("subject")
	assert.Nil(t, err)
	assert.Equal(t, expectedCookie, cookie)
}

func TestNewAuth(t *testing.T) {
	// Test case: Successful creation of Auth
	validator := &mockValidator{}
	logger := zaptest.NewLogger(t)
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NotNil(t, auth)
	assert.Nil(t, err)
	assert.Equal(t, DefaultContextKey, auth.contextKey)
	assert.Equal(t, logger, auth.Logger())

	// Test case: Missing context key
	auth, err = NewAuth(validator, WithContextKey(""))
	assert.Nil(t, auth)
	assert.Equal(t, ErrEmptyContextKey, err)

	// Test case: Missing logger
	auth, err = NewAuth(validator)
	assert.Nil(t, auth)
	assert.Equal(t, ErrEmptyLogger, err)
}
