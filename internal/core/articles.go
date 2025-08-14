package core

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/models"
	"strings"
	"time"
)

var ErrDuplicatedSlug = xerrors.Message("Duplicate slug")

func (c *Core) CreateArticle(article *models.Article) (*models.Article, error) {

	const insertSQL = `
		INSERT INTO articles (slug,title,description,body,created_at,updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id,slug,title,description,body,created_at,updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	modelArticle := &models.Article{}

	err := c.db.QueryRowContext(ctx, insertSQL, article.Slug, article.Title, article.Description, article.Body, time.Now(), time.Now()).
		Scan(&modelArticle.ID, &modelArticle.Slug, &modelArticle.Title, &modelArticle.Description, &modelArticle.Body, &modelArticle.CreatedAt, &modelArticle.UpdatedAt)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `duplicate key value violates unique constraint`):
			return nil, xerrors.New(ErrDuplicatedSlug)
		default:
			return nil, xerrors.New(err)
		}
	}
	return modelArticle, nil
}

func (c *Core) IsFavouriteArticleByUser(articleId int64, user *auth.User) (bool, error) {
	if user == nil {
		return false, nil
	}

	const selectSQL = `
		SELECT EXISTS( 
			SELECT 1 FROM favourite_articles WHERE user_id = $1 and article_id = $2
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var isFavourite bool
	err := c.db.QueryRowContext(ctx, selectSQL, user.ID, articleId).Scan(&isFavourite)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // No record found, not a favourite
		}
		return false, xerrors.New(err)
	}
	return isFavourite, nil
}

func (c *Core) FavouriteArticleCount(articleId int64) (int64, error) {
	const selectSQL = `
		SELECT COUNT(*) FROM favourite_articles WHERE article_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var favouriteArticleCount int64
	err := c.db.QueryRowContext(ctx, selectSQL, articleId).Scan(&favouriteArticleCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, xerrors.New(err)
	}
	return favouriteArticleCount, nil
}

func (c *Core) CreateSlug(title string) string {
	slug := strings.ToLower(title)

	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove common punctuation
	replacements := []string{".", ",", "!", "?", ":", ";", "'", "\"", "(", ")", "[", "]", "{", "}", "/", "\\"}
	for _, char := range replacements {
		slug = strings.ReplaceAll(slug, char, "")
	}

	// Replace multiple consecutive hyphens with single hyphen
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	slug = strings.Trim(slug, "-")

	return slug
}
