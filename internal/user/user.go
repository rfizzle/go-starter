// Package user is an example domain wired end to end through the starter's SQL
// stack: sqlc-generated typed queries for the fixed ~90% of access
// (internal/storage/postgres/db), squirrel for the dynamic ~10% (filtered
// listing), and a Querier interface so the connection pool and a pgx.Tx are
// interchangeable for transaction propagation.
//
// It is intentionally small: copy it as the template for real domains.
package user

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rfizzle/go-starter/internal/storage/postgres/db"
)

// User is the domain representation of a user. It is decoupled from the
// sqlc-generated db.User (which exposes pgtype values) so handlers and business
// logic work with plain Go types.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// fromDB maps a sqlc row to the domain type. The password hash is deliberately
// dropped here so it never escapes the storage boundary.
func fromDB(row db.User) User {
	return User{
		ID:        uuidToString(row.ID),
		Email:     row.Email,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// uuidToString renders a pgtype.UUID as the canonical 8-4-4-4-12 hex form.
// Returns "" when the value is NULL/unset.
func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	b := id.Bytes
	const hextable = "0123456789abcdef"
	// 32 hex digits + 4 dashes.
	buf := make([]byte, 36)
	for i, j := 0, 0; i < len(b); i++ {
		if j == 8 || j == 13 || j == 18 || j == 23 {
			buf[j] = '-'
			j++
		}
		buf[j] = hextable[b[i]>>4]
		buf[j+1] = hextable[b[i]&0x0f]
		j += 2
	}
	return string(buf)
}

// parseUUID converts a canonical UUID string into a pgtype.UUID for use as a
// query argument. An empty or malformed string yields an invalid (NULL) UUID.
func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if s == "" {
		return id, nil
	}
	if err := id.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}
