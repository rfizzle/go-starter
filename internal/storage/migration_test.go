package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationStatus_Values(t *testing.T) {
	tests := []struct {
		name     string
		status   MigrationStatus
		expected int
	}{
		{
			name:     "NoMigrationsCompleted",
			status:   NoMigrationsCompleted,
			expected: 0,
		},
		{
			name:     "SomeMigrationsCompleted",
			status:   SomeMigrationsCompleted,
			expected: 1,
		},
		{
			name:     "AllMigrationsCompleted",
			status:   AllMigrationsCompleted,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, int(tt.status))
		})
	}
}

func TestMigrationStatus_TypeSafety(t *testing.T) {
	// Test that MigrationStatus is a distinct type
	var status = NoMigrationsCompleted
	assert.IsType(t, MigrationStatus(0), status)

	// Test that it can be compared
	assert.Equal(t, NoMigrationsCompleted, status)
	assert.NotEqual(t, SomeMigrationsCompleted, status)
	assert.NotEqual(t, AllMigrationsCompleted, status)
}

func TestMigrationStatus_Ordering(t *testing.T) {
	// Test that the migration statuses have logical ordering
	assert.True(t, NoMigrationsCompleted < SomeMigrationsCompleted)
	assert.True(t, SomeMigrationsCompleted < AllMigrationsCompleted)
	assert.True(t, NoMigrationsCompleted < AllMigrationsCompleted)
}

func TestMigrationStatus_Constants(t *testing.T) {
	// Test that constants are immutable and have expected values
	const expectedNoMigrations = 0
	const expectedSomeMigrations = 1
	const expectedAllMigrations = 2

	assert.Equal(t, expectedNoMigrations, int(NoMigrationsCompleted))
	assert.Equal(t, expectedSomeMigrations, int(SomeMigrationsCompleted))
	assert.Equal(t, expectedAllMigrations, int(AllMigrationsCompleted))
}
