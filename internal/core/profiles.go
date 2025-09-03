package core

import (
	"context"
	"database/sql"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/models"
	"strings"
)

var (
	UserIsAlreadyFollowed = xerrors.Message("User is already followed")
	UserIsNotFollowed     = xerrors.Message("User is not followed")
)

// todo: use one sql query to fetch user and following status
func (c *Core) GetProfile(ctx context.Context, username string) (*models.Profile, error) {

	const queryFollowing = `
		SELECT EXISTS (
			SELECT 1 FROM followers WHERE follower_id = $1
		)
	`

	// Fetch user info
	profile := &models.Profile{}
	user, err := c.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	profile.ID = user.ID
	profile.Username = user.Username
	profile.Bio = user.Bio
	profile.Image = user.Image

	// Check following status
	isFollowing, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, ctx, queryFollowing, func(rows *sql.Rows) (bool, error) {
		var isFollowing bool
		if err := rows.Scan(&isFollowing); err != nil {
			return false, xerrors.New(err)
		}
		return isFollowing, nil
	}, profile.ID)

	if err != nil {
		return nil, xerrors.New(err)
	}

	profile.Following = isFollowing

	return profile, nil
}

func (c *Core) GetFollowingUserList(ctx context.Context, username string) ([]*auth.User, error) {
	user, err := c.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	queryFollowing := `
		SELECT EXISTS (
			SELECT u.id, u.email, u.username, u.password, u.bio, u.image 
			FROM users as u join followers f on u.id = f.user_id  
			WHERE follower_id = $1
		)
	`
	queryResultList, err := databaseutils.ExecuteQuery(c.sqlTemplate, ctx, queryFollowing, func(rows *sql.Rows) (*auth.User, error) {
		var tempUser = &auth.User{}

		if err := rows.Scan(&tempUser.ID,
			&tempUser.Email,
			&tempUser.Username,
			&tempUser.Password,
			&tempUser.Bio,
			&tempUser.Image); err != nil {
			return nil, xerrors.New(err)
		}
		return tempUser, nil
	}, user.ID)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return queryResultList, nil
}

func (c *Core) FollowUser(ctx context.Context, followerUser auth.User, followeeUserName string) (*models.Profile, error) {

	followeeUser, err := c.GetUserByUsername(ctx, followeeUserName)
	if err != nil {
		return nil, xerrors.New(err)
	}

	insertSql := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
		RETURNING user_id, follower_id
	`

	args := []interface{}{followeeUser.ID, followerUser.ID}

	_, err = databaseutils.ExecuteSingleQuery(c.sqlTemplate, ctx, insertSql, func(rows *sql.Rows) (bool, error) {
		var followerID, followeeID int64
		if err := rows.Scan(&followerID, &followeeID); err != nil {
			return false, xerrors.New(err)
		}
		return true, nil
	}, args)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), `duplicate key value violates unique constraint`):
			return nil, xerrors.New(UserIsAlreadyFollowed)
		default:
			return nil, xerrors.New(err)
		}
	}

	profile, err := c.GetProfile(ctx, followerUser.Username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	return profile, nil
}

func (c *Core) UnfollowUser(ctx context.Context, followerUser auth.User, followeeUserName string) (*models.Profile, error) {
	followeeUser, err := c.GetUserByUsername(ctx, followeeUserName)
	if err != nil {
		return nil, xerrors.New(err)
	}

	deleteSql := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`

	rowsAffected, err := databaseutils.ExecuteDeleteQuery(c.sqlTemplate, ctx, deleteSql, followeeUser.ID, followerUser.ID)

	if err != nil {
		return nil, xerrors.New(err)
	}

	if rowsAffected == 0 {
		return nil, xerrors.New(UserIsNotFollowed)
	}

	profile, err := c.GetProfile(ctx, followerUser.Username)
	if err != nil {
		return nil, xerrors.New(err)
	}

	return profile, nil
}
