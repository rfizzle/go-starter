package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// mockDBConn implements dbconn interface for testing
type mockDBConn struct {
	BeginFunc    func(ctx context.Context) (pgx.Tx, error)
	ExecFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	PingFunc     func(ctx context.Context) error
}

func (m *mockDBConn) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc(ctx)
	}
	return nil, nil
}

func (m *mockDBConn) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockDBConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockDBConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *mockDBConn) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
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

func TestClient_Name(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	client, err := Open(ctx, WithMockPool(mock))
	assert.NoError(t, err)

	assert.Equal(t, "postgresql", client.Name())
}

func TestClient_Ping(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		mockConn *mockDBConn
		timeout  int
		wantErr  bool
	}{
		{
			name: "successful ping",
			mockConn: &mockDBConn{
				PingFunc: func(ctx context.Context) error {
					return nil
				},
			},
			timeout: 15,
			wantErr: false,
		},
		{
			name: "ping timeout",
			mockConn: &mockDBConn{
				PingFunc: func(ctx context.Context) error {
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
			mockConn: &mockDBConn{
				PingFunc: func(ctx context.Context) error {
					return errors.New("connection refused")
				},
			},
			timeout: 15,
			wantErr: true,
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
				closeFunc:   func(ctx context.Context) error { return nil },
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

func TestClient_Close(t *testing.T) {
	tests := []struct {
		name      string
		closeFunc func(ctx context.Context) error
		wantErr   bool
	}{
		{
			name: "successful close",
			closeFunc: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "close error",
			closeFunc: func(ctx context.Context) error {
				return errors.New("failed to close connection")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				logger:    zaptest.NewLogger(t),
				db:        &mockDBConn{},
				closeFunc: tt.closeFunc,
			}

			err := client.Close()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Logger(t *testing.T) {
	logger := zaptest.NewLogger(t)

	client := &Client{
		logger: logger,
	}

	result := client.Logger()
	assert.Equal(t, logger, result)
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
				db:        &mockDBConn{},
				closeFunc: func(ctx context.Context) error { return nil },
				logger:    zaptest.NewLogger(nil),
			},
			wantErr: false,
		},
		{
			name: "nil db",
			client: &Client{
				db:        nil,
				closeFunc: func(ctx context.Context) error { return nil },
				logger:    zaptest.NewLogger(nil),
			},
			wantErr:   true,
			errString: "db is nil",
		},
		{
			name: "nil closeFunc",
			client: &Client{
				db:        &mockDBConn{},
				closeFunc: nil,
				logger:    zaptest.NewLogger(nil),
			},
			wantErr:   true,
			errString: "closeFunc is nil",
		},
		{
			name: "nil logger",
			client: &Client{
				db:        &mockDBConn{},
				closeFunc: func(ctx context.Context) error { return nil },
				logger:    nil,
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

	// Test close
	err = client.Close()
	assert.NoError(t, err)

	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
