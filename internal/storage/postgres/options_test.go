package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestWithURI(t *testing.T) {
	ctx := context.Background()
	provider := &Provider{}

	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "invalid URI",
			uri:     "invalid://uri",
			wantErr: true,
		},
		// Note: We can't easily test valid URIs without a real database
		// In integration tests, you would test with a real database URI
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := WithURI(ctx, tt.uri)
			err := option(provider)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider.db)
				assert.NotNil(t, provider.closeFunc)
			}
		})
	}
}

func TestWithConnection(t *testing.T) {
	// Skip this test as it requires actual pgx.Conn
	t.Skip("Requires actual pgx.Conn implementation")
}

func TestWithPool(t *testing.T) {
	// Skip this test as it requires actual pgxpool.Pool
	t.Skip("Requires actual pgxpool.Pool implementation")
}

func TestWithMockPool(t *testing.T) {
	ctx := context.Background()
	provider := &Provider{}

	// Create pgxmock
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	option := WithMockPool(mock)
	err = option(provider)

	assert.NoError(t, err)
	assert.Equal(t, mock, provider.db)
	assert.NotNil(t, provider.closeFunc)

	// Test the close function doesn't panic
	err = provider.closeFunc(ctx)
	assert.NoError(t, err)
}

func TestWithPingTimeout(t *testing.T) {
	provider := &Provider{}

	tests := []struct {
		name            string
		timeout         int
		expectedTimeout time.Duration
	}{
		{
			name:            "positive timeout",
			timeout:         30,
			expectedTimeout: 30 * time.Second,
		},
		{
			name:            "zero timeout",
			timeout:         0,
			expectedTimeout: 0,
		},
		{
			name:            "large timeout",
			timeout:         300,
			expectedTimeout: 300 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := WithPingTimeout(tt.timeout)
			err := option(provider)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTimeout, provider.pingTimeout)
		})
	}
}

func TestWithLogger(t *testing.T) {
	provider := &Provider{}

	tests := []struct {
		name    string
		logger  *zap.Logger
		wantErr bool
	}{
		{
			name:    "valid logger",
			logger:  zaptest.NewLogger(t),
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := WithLogger(tt.logger)
			err := option(provider)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "logger cannot be nil")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.logger, provider.logger)
			}
		})
	}
}

func TestOptionsChaining(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	// Test that multiple options can be chained together
	provider, err := Open(ctx,
		WithMockPool(mock),
		WithLogger(logger),
		WithPingTimeout(45),
	)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, mock, provider.db)
	assert.Equal(t, logger, provider.logger)
	assert.Equal(t, 45*time.Second, provider.pingTimeout)
}
