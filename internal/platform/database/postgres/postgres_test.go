package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type mockPinger struct {
	pingFunc func(context.Context) error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func TestOpen(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name      string
		options   []Option
		wantErr   bool
		errString string
	}{
		{
			name: "successful creation with mock pool",
			options: func() []Option {
				mock, _ := pgxmock.NewPool()
				return []Option{
					WithMockPool(mock),
					WithLogger(logger),
				}
			}(),
			wantErr: false,
		},
		{
			name: "successful creation with ping timeout",
			options: func() []Option {
				mock, _ := pgxmock.NewPool()
				return []Option{
					WithMockPool(mock),
					WithLogger(logger),
					WithPingTimeout(30),
				}
			}(),
			wantErr: false,
		},
		{
			name: "validation error - nil db",
			options: []Option{
				WithLogger(logger),
			},
			wantErr:   true,
			errString: "db is nil",
		},
		{
			name: "validation error - nil logger",
			options: func() []Option {
				mock, _ := pgxmock.NewPool()
				return []Option{
					WithMockPool(mock),
					WithLogger(nil),
				}
			}(),
			wantErr:   true,
			errString: "logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := Open(ctx, tt.options...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		mockConn *mockPinger
		timeout  int
		wantErr  bool
	}{
		{
			name: "successful ping",
			mockConn: &mockPinger{
				pingFunc: func(ctx context.Context) error {
					return nil
				},
			},
			timeout: 15,
			wantErr: false,
		},
		{
			name: "ping timeout",
			mockConn: &mockPinger{
				pingFunc: func(ctx context.Context) error {
					select {
					case <-time.After(100 * time.Millisecond):
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				},
			},
			timeout: 1, // Special value - converted to 1ms in test
			wantErr: true,
		},
		{
			name: "ping database error",
			mockConn: &mockPinger{
				pingFunc: func(ctx context.Context) error {
					return errors.New("connection refused")
				},
			},
			timeout: 15,
			wantErr: true,
		},
		{
			name: "zero timeout defers to caller context",
			mockConn: &mockPinger{
				pingFunc: func(ctx context.Context) error {
					if _, ok := ctx.Deadline(); ok {
						return errors.New("unexpected client deadline")
					}
					return nil
				},
			},
			timeout: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var timeout time.Duration
			if tt.timeout == 1 { // Special case for timeout test
				timeout = 1 * time.Millisecond
			} else {
				timeout = time.Duration(tt.timeout) * time.Second
			}

			client := &Client{
				logger:      zaptest.NewLogger(t),
				db:          tt.mockConn,
				pingTimeout: timeout,
			}

			err := client.Ping(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_validate(t *testing.T) {
	tests := []struct {
		name      string
		client    *Client
		wantErr   bool
		errString string
	}{
		{
			name: "valid client",
			client: &Client{
				db:     &mockPinger{},
				logger: zaptest.NewLogger(nil),
			},
			wantErr: false,
		},
		{
			name: "nil db",
			client: &Client{
				db:     nil,
				logger: zaptest.NewLogger(nil),
			},
			wantErr:   true,
			errString: "db is nil",
		},
		{
			name: "nil logger",
			client: &Client{
				db:     &mockPinger{},
				logger: nil,
			},
			wantErr:   true,
			errString: "logger is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_IntegrationWithPgxMock(t *testing.T) {
	ctx := context.Background()

	// Create pgxmock
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	// Create client with mock
	client, err := Open(ctx, WithMockPool(mock), WithLogger(zaptest.NewLogger(t)))
	assert.NoError(t, err)

	// Test ping
	mock.ExpectPing()
	err = client.Ping(ctx)
	assert.NoError(t, err)

	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
