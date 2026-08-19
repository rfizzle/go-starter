// Package user implements PostgreSQL persistence for the example user business
// module. Shared user data, inputs, and errors live in internal/domain;
// SQL and pgx types remain inside this adapter.
//
// It is intentionally small: copy it as the template for real domains.
package user

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rfizzle/go-starter/internal/domain"
	"github.com/rfizzle/go-starter/internal/platform/database/postgres/db"
)

// fromDB maps a sqlc row to the domain type. The password hash is deliberately
// dropped here so it never escapes the storage boundary.
func fromDB(row db.User) domain.User {
	return domain.User{
		ID:        uuidToString(row.ID),
		Email:     row.Email,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// credentialFromDB maps a sqlc row to the authentication-only projection.
// Unlike fromDB, it retains the password hash: this is the sole conversion
// permitted to do so, and only Repository.CredentialByEmail calls it.
func credentialFromDB(row db.User) domain.Credential {
	return domain.Credential{
		UserID:       uuidToString(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
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
