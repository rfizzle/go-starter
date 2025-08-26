package postgres

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

// Test struct for scanning
type testUser struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestInternalScanner_Integration(t *testing.T) {
	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	t.Run("ScanOne success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "name"}).
			AddRow(1, "John Doe")

		mock.ExpectQuery("SELECT id, name FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		resultRows, err := mock.Query(ctx, "SELECT id, name FROM users WHERE id = $1", 1)
		assert.NoError(t, err)

		scanner := &internalScanner{rows: resultRows}
		var user testUser
		err = scanner.ScanOne(&user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "John Doe", user.Name)
	})

	t.Run("ScanAll success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "name"}).
			AddRow(1, "John Doe").
			AddRow(2, "Jane Smith")

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		resultRows, err := mock.Query(ctx, "SELECT id, name FROM users")
		assert.NoError(t, err)

		scanner := &internalScanner{rows: resultRows}
		var users []testUser
		err = scanner.ScanAll(&users)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "John Doe", users[0].Name)
		assert.Equal(t, "Jane Smith", users[1].Name)
	})

	t.Run("ScanAll empty result", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery("SELECT id, name FROM users WHERE 1=0").
			WillReturnRows(rows)

		resultRows, err := mock.Query(ctx, "SELECT id, name FROM users WHERE 1=0")
		assert.NoError(t, err)

		scanner := &internalScanner{rows: resultRows}
		var users []testUser
		err = scanner.ScanAll(&users)
		assert.NoError(t, err)
		assert.Len(t, users, 0)
	})

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
