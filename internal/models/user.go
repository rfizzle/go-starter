package models

type User struct {
	ID           string `json:"-" db:"id"`
	Email        string `json:"-" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
	FirstName    string `json:"-" db:"first_name"`
	LastName     string `json:"-" db:"last_name"`
}
