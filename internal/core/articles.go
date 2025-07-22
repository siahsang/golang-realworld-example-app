package core

import (
	"context"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/models"
	"time"
)

func (c *Core) CreateArticle(username string) (*models.Profile, error) {

	const queryFollowing = `
		SELECT EXISTS (
			SELECT 1 FROM followers WHERE follower_id = $1
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Fetch user info
	profile := &models.Profile{}
	user, err := c.GetUserByUsername(username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	profile.ID = user.ID
	profile.Username = user.Username
	profile.Bio = user.Bio
	profile.Image = user.Image

	// Check following status
	err = c.db.QueryRowContext(ctx, queryFollowing, profile.ID).Scan(&profile.Following)
	if err != nil {
		return nil, xerrors.New(err)
	}

	return profile, nil
}
