package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/utils/database"
	"github.com/siahsang/blog/internal/utils/stringutils"
	"strings"
	"time"
)

var (
	ErrDuplicateEmail    = xerrors.Message("Duplicate email")
	ErrDuplicateUsername = xerrors.Message("Duplicate username")
	NoRecordFound        = xerrors.Message("No record found")
)

func (c *Core) CreateNewUser(user *auth.User) error {
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
			return xerrors.New(ErrDuplicateEmail)
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return xerrors.New(ErrDuplicateUsername)
		default:
			return xerrors.New(err)
		}
	}

	return nil
}

func (c *Core) GetUserByEmail(email string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password, bio, image
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
		&user.Bio,
		&user.Image,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, xerrors.New(NoRecordFound)
		default:
			return nil, xerrors.New(err)
		}
	}

	return &user, nil
}

func (c *Core) GetUserByUsername(username string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password, bio, image
		FROM users
		WHERE username = $1
	`

	var user auth.User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Bio,
		&user.Image,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, xerrors.New(NoRecordFound)
		default:
			return nil, xerrors.New(err)
		}
	}

	return &user, nil
}

func (c *Core) GetUsersByIdList(userIdList []int64) ([]*auth.User, error) {
	placeholders, args := stringutils.INCluse(userIdList)
	query := fmt.Sprintf(`
		SELECT id, email, username, password, bio, image
		FROM users
		WHERE id  in (%s)
	`, strings.Join(placeholders, ", "))

	queryResultList, err := database.ExecuteQuery(c.sqlTemplate, query, func(rows *sql.Rows) (*auth.User, error) {
		var user *auth.User

		if err := rows.Scan(&user.ID,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Bio,
			&user.Image); err != nil {
			return nil, xerrors.New(err)
		}
		return user, nil
	}, args...)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return queryResultList, nil
}

func (c *Core) UpdateUser(user *auth.User) (*auth.User, error) {
	query := `
		UPDATE users
		SET bio = $1,image= $2
		WHERE id = $3
		RETURNING id, email, username, bio, image 
		
	`

	var returningUser auth.User
	args := []interface{}{user.Bio, user.Image, user.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.db.QueryRowContext(ctx, query, args...).Scan(
		&returningUser.ID,
		&returningUser.Email,
		&returningUser.Username,
		&returningUser.Bio,
		&returningUser.Image,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, xerrors.New(NoRecordFound)
		default:
			return nil, xerrors.New(err)
		}
	}

	c.log.Info("User updated Successfully", "user_id", returningUser.ID, "email", returningUser.Email)
	return &returningUser, nil
}
