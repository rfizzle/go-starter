package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockScanner implements the Scanner interface for testing
type MockScanner struct {
	ScanOneFunc func(dst interface{}) error
	ScanAllFunc func(dst interface{}) error
}

func (m *MockScanner) ScanOne(dst interface{}) error {
	if m.ScanOneFunc != nil {
		return m.ScanOneFunc(dst)
	}
	return nil
}

func (m *MockScanner) ScanAll(dst interface{}) error {
	if m.ScanAllFunc != nil {
		return m.ScanAllFunc(dst)
	}
	return nil
}

// MockClient implements the Client interface for testing
type MockClient struct {
	QueryFunc   func(sql string, args ...interface{}) (Scanner, error)
	ExecuteFunc func(sql string, args ...interface{}) error
}

func (m *MockClient) Query(sql string, args ...interface{}) (Scanner, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(sql, args...)
	}
	return &MockScanner{}, nil
}

func (m *MockClient) Execute(sql string, args ...interface{}) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(sql, args...)
	}
	return nil
}

func TestMockScanner_ScanOne(t *testing.T) {
	tests := []struct {
		name    string
		scanner *MockScanner
		dst     interface{}
		wantErr bool
	}{
		{
			name: "successful scan",
			scanner: &MockScanner{
				ScanOneFunc: func(dst interface{}) error {
					return nil
				},
			},
			dst:     &struct{}{},
			wantErr: false,
		},
		{
			name: "scan error",
			scanner: &MockScanner{
				ScanOneFunc: func(dst interface{}) error {
					return errors.New("scan failed")
				},
			},
			dst:     &struct{}{},
			wantErr: true,
		},
		{
			name:    "nil function defaults to no error",
			scanner: &MockScanner{},
			dst:     &struct{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scanner.ScanOne(tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockScanner_ScanAll(t *testing.T) {
	tests := []struct {
		name    string
		scanner *MockScanner
		dst     interface{}
		wantErr bool
	}{
		{
			name: "successful scan all",
			scanner: &MockScanner{
				ScanAllFunc: func(dst interface{}) error {
					return nil
				},
			},
			dst:     &[]struct{}{},
			wantErr: false,
		},
		{
			name: "scan all error",
			scanner: &MockScanner{
				ScanAllFunc: func(dst interface{}) error {
					return errors.New("scan all failed")
				},
			},
			dst:     &[]struct{}{},
			wantErr: true,
		},
		{
			name:    "nil function defaults to no error",
			scanner: &MockScanner{},
			dst:     &[]struct{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scanner.ScanAll(tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockClient_Query(t *testing.T) {
	tests := []struct {
		name    string
		client  *MockClient
		sql     string
		args    []interface{}
		wantErr bool
	}{
		{
			name: "successful query",
			client: &MockClient{
				QueryFunc: func(sql string, args ...interface{}) (Scanner, error) {
					return &MockScanner{}, nil
				},
			},
			sql:     "SELECT * FROM users",
			args:    []interface{}{1},
			wantErr: false,
		},
		{
			name: "query error",
			client: &MockClient{
				QueryFunc: func(sql string, args ...interface{}) (Scanner, error) {
					return nil, errors.New("query failed")
				},
			},
			sql:     "SELECT * FROM users WHERE id = ?",
			args:    []interface{}{1},
			wantErr: true,
		},
		{
			name:    "nil function returns mock scanner",
			client:  &MockClient{},
			sql:     "SELECT * FROM users",
			args:    []interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := tt.client.Query(tt.sql, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, scanner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scanner)
			}
		})
	}
}

func TestMockClient_Execute(t *testing.T) {
	tests := []struct {
		name    string
		client  *MockClient
		sql     string
		args    []interface{}
		wantErr bool
	}{
		{
			name: "successful execute",
			client: &MockClient{
				ExecuteFunc: func(sql string, args ...interface{}) error {
					return nil
				},
			},
			sql:     "INSERT INTO users (name) VALUES (?)",
			args:    []interface{}{"John"},
			wantErr: false,
		},
		{
			name: "execute error",
			client: &MockClient{
				ExecuteFunc: func(sql string, args ...interface{}) error {
					return errors.New("execute failed")
				},
			},
			sql:     "UPDATE users SET name = ? WHERE id = ?",
			args:    []interface{}{"Jane", 1},
			wantErr: true,
		},
		{
			name:    "nil function returns no error",
			client:  &MockClient{},
			sql:     "DELETE FROM users WHERE id = ?",
			args:    []interface{}{1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Execute(tt.sql, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
