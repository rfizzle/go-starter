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

func TestWithPool(t *testing.T) {
	// Skip this test as it requires actual pgxpool.Pool
	t.Skip("Requires actual pgxpool.Pool implementation")
}

func TestWithMockPool(t *testing.T) {
	client := &Client{}

	// Create pgxmock
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	option := WithMockPool(mock)
	err = option(client)

	assert.NoError(t, err)
	assert.Equal(t, mock, client.db)
}

func TestWithPingTimeout(t *testing.T) {
	client := &Client{}

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
			err := option(client)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTimeout, client.pingTimeout)
		})
	}
}

func TestWithLogger(t *testing.T) {
	client := &Client{}

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
			err := option(client)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "logger cannot be nil")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.logger, client.logger)
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
	client, err := Open(ctx,
		WithMockPool(mock),
		WithLogger(logger),
		WithPingTimeout(45),
	)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, mock, client.db)
	assert.Equal(t, logger, client.logger)
	assert.Equal(t, 45*time.Second, client.pingTimeout)
}
