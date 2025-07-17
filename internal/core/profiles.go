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
	const queryUserInfo = `
		SELECT id, username, bio, image
		FROM users
		WHERE username = $1
	`

	const queryFollowing = `
		SELECT EXISTS (
			SELECT 1 FROM followers WHERE follower_id = $1
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Fetch user info
	profile := &models.Profile{}
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

	// Check following status
	err = c.db.QueryRowContext(ctx, queryFollowing, profile.ID).Scan(&profile.Following)
	if err != nil {
		return nil, xerrors.New(err)
	}

	return profile, nil
}
