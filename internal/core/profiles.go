package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/models"
	"time"
)

func (c *Core) GetProfile(username string) (*models.Profile, error) {
	queryUserInfo := `
		SELECT id, username, bio, image
		FROM users
		WHERE username = $1
	`

	queryFollowing := `SELECT EXISTS (SELECT user_id FROM followers WHERE follower_id = $1)`

	var profile models.Profile

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.db.QueryRowContext(ctx, queryUserInfo, username).Scan(
		&profile.ID,
		&profile.Username,
		&profile.Bio,
		&profile.Image,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, xerrors.New(NoRecordFound)
		}
		return nil, xerrors.New(err)
	}

	err = c.db.QueryRowContext(ctx, queryFollowing, profile.ID).Scan(
		&profile.Following,
	)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return &profile, nil
}
