package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/utils/stringutils"
	"strings"
)

var (
	ErrDuplicateEmail    = xerrors.Message("Duplicate email")
	ErrDuplicateUsername = xerrors.Message("Duplicate username")
	NoRecordFound        = xerrors.Message("No record found")
)

func (c *Core) CreateNewUser(context context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id
`
	args := []any{user.Username, user.Email, user.Password}
	_, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*auth.User, error) {
		if err := rows.Scan(&user.ID); err != nil {
			return nil, xerrors.New(err)
		}
		return user, nil
	}, args...)

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

func (c *Core) GetUserByEmail(context context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password, bio, image
		FROM users
		WHERE email = $1
	`

	user, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*auth.User, error) {
		var user = &auth.User{}

		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Bio,
			&user.Image,
		); err != nil {
			return nil, xerrors.New(err)
		}
		return user, nil
	}, email)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, xerrors.New(NoRecordFound)
		default:
			return nil, xerrors.New(err)
		}
	}

	return user, nil
}

func (c *Core) GetUserByUsername(context context.Context, username string) (*auth.User, error) {
	query := `
		SELECT id, email, username, password, bio, image
		FROM users
		WHERE username = $1
	`

	user, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*auth.User, error) {
		var user = &auth.User{}

		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Bio,
			&user.Image,
		); err != nil {
			return nil, xerrors.New(err)
		}
		return user, nil
	}, username)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, xerrors.New(NoRecordFound)
		default:
			return nil, xerrors.New(err)
		}
	}

	return user, nil
}

func (c *Core) GetUsersByIdList(context context.Context, userIdList []int64) ([]*auth.User, error) {
	if len(userIdList) == 0 {
		return []*auth.User{}, nil
	}

	placeholders, args := stringutils.INCluse(userIdList)
	query := fmt.Sprintf(`
		SELECT id, email, username, password, bio, image
		FROM users
		WHERE id in (%s)
	`, strings.Join(placeholders, ", "))

	queryResultList, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*auth.User, error) {
		var user = &auth.User{}

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

func (c *Core) UpdateUser(context context.Context, user *auth.User) (*auth.User, error) {
	query := `
		UPDATE users
		SET bio = $1,image= $2
		WHERE id = $3
		RETURNING id, email, username, bio, image 
		
	`

	args := []any{user.Bio, user.Image, user.ID}
	returningUser, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*auth.User, error) {
		var user = &auth.User{}

		if err := rows.Scan(&user.ID,
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Bio,
			&user.Image); err != nil {
			return nil, xerrors.New(err)
		}
		return user, nil
	}, args...)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, xerrors.New(NoRecordFound)
		default:
			return nil, xerrors.New(err)
		}
	}

	c.log.Info("User updated Successfully", "user_id", returningUser.ID, "email", returningUser.Email)
	return returningUser, nil
}
