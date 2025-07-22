package core

import (
	"context"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/models"
	"strings"
	"time"
)

var (
	UserIsAlreadyFollowed = xerrors.Message("User is already followed")
	UserIsNotFollowed     = xerrors.Message("User is not followed")
)

// todo: use one sql query to fetch user and following status
func (c *Core) GetProfile(username string) (*models.Profile, error) {

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

func (c *Core) FollowUser(followerUser auth.User, followeeUserName string) (*models.Profile, error) {

	followeeUser, err := c.GetUserByUsername(followeeUserName)
	if err != nil {
		return nil, xerrors.New(err)
	}

	insertSql := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
		RETURNING user_id, follower_id
	`

	var followerID, followeeID int64
	args := []interface{}{followeeUser.ID, followerUser.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = c.db.QueryRowContext(ctx, insertSql, args...).Scan(&followerID, &followeeID)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), `duplicate key value violates unique constraint`):
			return nil, xerrors.New(UserIsAlreadyFollowed)
		default:
			return nil, xerrors.New(err)
		}
	}

	profile, err := c.GetProfile(followerUser.Username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	return profile, nil
}

func (c *Core) UnfollowUser(followerUser auth.User, followeeUserName string) (*models.Profile, error) {
	followeeUser, err := c.GetUserByUsername(followeeUserName)
	if err != nil {
		return nil, xerrors.New(err)
	}

	deleteSql := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.db.ExecContext(ctx, deleteSql, followeeUser.ID, followerUser.ID)
	if err != nil {
		return nil, xerrors.New(err)
	}
	affected, err := result.RowsAffected()

	if err != nil {
		return nil, xerrors.New(err)
	}

	if affected == 0 {
		return nil, xerrors.New(UserIsNotFollowed)
	}

	profile, err := c.GetProfile(followerUser.Username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	return profile, nil
}
