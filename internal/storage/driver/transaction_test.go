package driver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockTx implements the Tx interface for testing
type MockTx struct {
	InsertManyFunc func(queries map[string][]interface{}) error
	QueryFunc      func(sql string, args ...interface{}) (Scanner, error)
	ExecuteFunc    func(sql string, args ...interface{}) error
}

func (m *MockTx) InsertMany(queries map[string][]interface{}) error {
	if m.InsertManyFunc != nil {
		return m.InsertManyFunc(queries)
	}
	return nil
}

func (m *MockTx) Query(sql string, args ...interface{}) (Scanner, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(sql, args...)
	}
	return &MockScanner{}, nil
}

func (m *MockTx) Execute(sql string, args ...interface{}) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(sql, args...)
	}
	return nil
}

func TestMockTx_InsertMany(t *testing.T) {
	tests := []struct {
		name    string
		tx      *MockTx
		queries map[string][]interface{}
		wantErr bool
	}{
		{
			name: "successful batch insert",
			tx: &MockTx{
				InsertManyFunc: func(queries map[string][]interface{}) error {
					return nil
				},
			},
			queries: map[string][]interface{}{
				"INSERT INTO users (name, email) VALUES (?, ?)":     {"John", "john@example.com"},
				"INSERT INTO profiles (user_id, bio) VALUES (?, ?)": {1, "Developer"},
			},
			wantErr: false,
		},
		{
			name: "batch insert error",
			tx: &MockTx{
				InsertManyFunc: func(queries map[string][]interface{}) error {
					return errors.New("batch insert failed")
				},
			},
			queries: map[string][]interface{}{
				"INSERT INTO users (name) VALUES (?)": {"Jane"},
			},
			wantErr: true,
		},
		{
			name:    "empty queries map",
			tx:      &MockTx{},
			queries: map[string][]interface{}{},
			wantErr: false,
		},
		{
			name:    "nil queries map",
			tx:      &MockTx{},
			queries: nil,
			wantErr: false,
		},
		{
			name: "default implementation success",
			tx:   &MockTx{},
			queries: map[string][]interface{}{
				"INSERT INTO test (value) VALUES (?)": {"test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.InsertMany(tt.queries)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockTx_Query(t *testing.T) {
	tests := []struct {
		name    string
		tx      *MockTx
		sql     string
		args    []interface{}
		wantErr bool
	}{
		{
			name: "successful transaction query",
			tx: &MockTx{
				QueryFunc: func(sql string, args ...interface{}) (Scanner, error) {
					return &MockScanner{}, nil
				},
			},
			sql:     "SELECT * FROM users WHERE id = ?",
			args:    []interface{}{1},
			wantErr: false,
		},
		{
			name: "transaction query error",
			tx: &MockTx{
				QueryFunc: func(sql string, args ...interface{}) (Scanner, error) {
					return nil, errors.New("query in transaction failed")
				},
			},
			sql:     "SELECT * FROM users",
			args:    []interface{}{},
			wantErr: true,
		},
		{
			name:    "default query implementation",
			tx:      &MockTx{},
			sql:     "SELECT COUNT(*) FROM users",
			args:    []interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := tt.tx.Query(tt.sql, tt.args...)
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

func TestMockTx_Execute(t *testing.T) {
	tests := []struct {
		name    string
		tx      *MockTx
		sql     string
		args    []interface{}
		wantErr bool
	}{
		{
			name: "successful transaction execute",
			tx: &MockTx{
				ExecuteFunc: func(sql string, args ...interface{}) error {
					return nil
				},
			},
			sql:     "UPDATE users SET name = ? WHERE id = ?",
			args:    []interface{}{"Updated Name", 1},
			wantErr: false,
		},
		{
			name: "transaction execute error",
			tx: &MockTx{
				ExecuteFunc: func(sql string, args ...interface{}) error {
					return errors.New("execute in transaction failed")
				},
			},
			sql:     "DELETE FROM users WHERE id = ?",
			args:    []interface{}{1},
			wantErr: true,
		},
		{
			name:    "default execute implementation",
			tx:      &MockTx{},
			sql:     "INSERT INTO logs (message) VALUES (?)",
			args:    []interface{}{"test log"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Execute(tt.sql, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockTx_InterfaceCompliance(t *testing.T) {
	// Ensure MockTx implements both Tx and Client interfaces
	var _ Tx = (*MockTx)(nil)
	var _ Client = (*MockTx)(nil)

	// Test that a MockTx can be used as both interfaces
	mockTx := &MockTx{}

	// Test as Client interface
	var client Client = mockTx
	scanner, err := client.Query("SELECT 1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, scanner)

	err = client.Execute("INSERT INTO test VALUES (?)", "value")
	assert.NoError(t, err)

	// Test as Tx interface
	var tx Tx = mockTx
	err = tx.InsertMany(map[string][]interface{}{
		"INSERT INTO test VALUES (?)": {"value"},
	})
	assert.NoError(t, err)
}
