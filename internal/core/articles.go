package core

import (
	"context"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/models"
	"strings"
	"time"
)

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
		return nil, xerrors.New(err)
	}

	return modelArticle, nil
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
