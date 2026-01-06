package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("Duplicate email")
	ErrEditConflict   = errors.New("Edit conflicts")
)

type User struct {
	ID         int64     `json:"id"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Role       string    `json:"role"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	version    int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextpassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextpassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextpassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextpassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextpassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

type UserModel struct {
	db *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
			INSERT INTO users (email, password_hash, role, is_verified)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, version
	`
	args := []any{user.Email, user.Password.hash, user.Role, user.IsVerified}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.version)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
