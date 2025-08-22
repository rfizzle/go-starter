package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	NameFunc       func() string
	PingFunc       func() error
	DisconnectFunc func() error
	LoggerFunc     func() *zap.Logger
	WithTxFunc     func(fn func(tx Tx) error) error
	QueryFunc      func(sql string, args ...interface{}) (Scanner, error)
	ExecuteFunc    func(sql string, args ...interface{}) error
}

func (m *MockProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock"
}

func (m *MockProvider) Ping() error {
	if m.PingFunc != nil {
		return m.PingFunc()
	}
	return nil
}

func (m *MockProvider) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	return nil
}

func (m *MockProvider) Logger() *zap.Logger {
	if m.LoggerFunc != nil {
		return m.LoggerFunc()
	}
	return zap.NewNop()
}

func (m *MockProvider) WithTx(fn func(tx Tx) error) error {
	if m.WithTxFunc != nil {
		return m.WithTxFunc(fn)
	}
	return fn(&MockTx{})
}

func (m *MockProvider) Query(sql string, args ...interface{}) (Scanner, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(sql, args...)
	}
	return &MockScanner{}, nil
}

func (m *MockProvider) Execute(sql string, args ...interface{}) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(sql, args...)
	}
	return nil
}

func TestMockProvider_Name(t *testing.T) {
	tests := []struct {
		name     string
		provider *MockProvider
		expected string
	}{
		{
			name: "custom name function",
			provider: &MockProvider{
				NameFunc: func() string {
					return "postgresql"
				},
			},
			expected: "postgresql",
		},
		{
			name:     "default name",
			provider: &MockProvider{},
			expected: "mock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.Name()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockProvider_Ping(t *testing.T) {
	tests := []struct {
		name     string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name: "successful ping",
			provider: &MockProvider{
				PingFunc: func() error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "ping error",
			provider: &MockProvider{
				PingFunc: func() error {
					return errors.New("connection failed")
				},
			},
			wantErr: true,
		},
		{
			name:     "default ping success",
			provider: &MockProvider{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.provider.Ping()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockProvider_Disconnect(t *testing.T) {
	tests := []struct {
		name     string
		provider *MockProvider
		wantErr  bool
	}{
		{
			name: "successful disconnect",
			provider: &MockProvider{
				DisconnectFunc: func() error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "disconnect error",
			provider: &MockProvider{
				DisconnectFunc: func() error {
					return errors.New("disconnect failed")
				},
			},
			wantErr: true,
		},
		{
			name:     "default disconnect success",
			provider: &MockProvider{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.provider.Disconnect()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockProvider_Logger(t *testing.T) {
	customLogger := zap.NewExample()

	tests := []struct {
		name     string
		provider *MockProvider
		expected *zap.Logger
	}{
		{
			name: "custom logger",
			provider: &MockProvider{
				LoggerFunc: func() *zap.Logger {
					return customLogger
				},
			},
			expected: customLogger,
		},
		{
			name:     "default nop logger",
			provider: &MockProvider{},
			expected: nil, // We'll check it's not nil but won't compare exact instance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.Logger()
			if tt.expected != nil {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMockProvider_WithTx(t *testing.T) {
	tests := []struct {
		name     string
		provider *MockProvider
		txFn     func(tx Tx) error
		wantErr  bool
	}{
		{
			name: "successful transaction",
			provider: &MockProvider{
				WithTxFunc: func(fn func(tx Tx) error) error {
					return fn(&MockTx{})
				},
			},
			txFn: func(tx Tx) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "transaction function error",
			provider: &MockProvider{
				WithTxFunc: func(fn func(tx Tx) error) error {
					return fn(&MockTx{})
				},
			},
			txFn: func(tx Tx) error {
				return errors.New("transaction failed")
			},
			wantErr: true,
		},
		{
			name: "withTx implementation error",
			provider: &MockProvider{
				WithTxFunc: func(fn func(tx Tx) error) error {
					return errors.New("begin transaction failed")
				},
			},
			txFn: func(tx Tx) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "default implementation",
			provider: &MockProvider{},
			txFn: func(tx Tx) error {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.provider.WithTx(tt.txFn)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
