package user

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rfizzle/go-starter/internal/platform/database/postgres/db"
)

func TestUUIDToString(t *testing.T) {
	t.Run("valid uuid renders canonical form", func(t *testing.T) {
		id := pgtype.UUID{
			Bytes: [16]byte{
				0x12, 0x34, 0x56, 0x78,
				0x9a, 0xbc,
				0xde, 0xf0,
				0x12, 0x34,
				0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
			},
			Valid: true,
		}
		assert.Equal(t, "12345678-9abc-def0-1234-56789abcdef0", uuidToString(id))
	})

	t.Run("invalid uuid renders empty string", func(t *testing.T) {
		assert.Equal(t, "", uuidToString(pgtype.UUID{Valid: false}))
	})
}

func TestParseUUIDRoundTrip(t *testing.T) {
	const canonical = "12345678-9abc-def0-1234-56789abcdef0"

	id, err := parseUUID(canonical)
	require.NoError(t, err)
	require.True(t, id.Valid)
	assert.Equal(t, canonical, uuidToString(id))
}

func TestParseUUID(t *testing.T) {
	t.Run("empty string yields invalid uuid without error", func(t *testing.T) {
		id, err := parseUUID("")
		require.NoError(t, err)
		assert.False(t, id.Valid)
	})

	t.Run("malformed string errors", func(t *testing.T) {
		_, err := parseUUID("not-a-uuid")
		assert.Error(t, err)
	})
}

func TestFromDB_DropsPasswordHash(t *testing.T) {
	row := db.User{
		Email:        "user@example.com",
		PasswordHash: "secret-hash",
		FirstName:    "Ada",
		LastName:     "Lovelace",
	}

	u := fromDB(row)

	assert.Equal(t, "user@example.com", u.Email)
	assert.Equal(t, "Ada", u.FirstName)
	assert.Equal(t, "Lovelace", u.LastName)
	// The domain type has no field for the password hash; this test documents
	// that the hash never crosses the storage boundary.
}
