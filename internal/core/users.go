package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"time"
)

var (
	ErrDuplicateEmail    = xerrors.Message("duplicate email")
	ErrDuplicateUsername = xerrors.Message("duplicate username")
	NoRecordFound        = xerrors.Message("record not found")
)

func (c *Core) Insert(user *auth.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id
`
	args := []interface{}{user.Username, user.Email, user.Password}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.db.QueryRowContext(ctx, query, args...).Scan(&user.ID)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return xerrors.New(err)
		}
	}

	return nil
}

func (c *Core) GetByEmail(email string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password
		FROM users
		WHERE email = $1
	`

	var user auth.User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, NoRecordFound
		default:
			return nil, xerrors.New(err)
		}
	}

	return &user, nil
}
