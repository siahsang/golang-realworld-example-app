package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/utils/database"
	"github.com/siahsang/blog/internal/utils/stringutils"
	"github.com/siahsang/blog/models"
	"strings"
	"time"
)

var ErrDuplicatedSlug = xerrors.Message("Duplicate slug")

func (c *Core) CreateArticle(article *models.Article) (*models.Article, error) {

	const insertSQL = `
		INSERT INTO articles (slug,title,description,body,created_at,updated_at,author_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id,slug,title,description,body,created_at,updated_at,author_id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	modelArticle := &models.Article{}

	err := c.db.QueryRowContext(ctx, insertSQL, article.Slug, article.Title, article.Description, article.Body, time.Now(), time.Now()).
		Scan(&modelArticle.ID, &modelArticle.Slug, &modelArticle.Title, &modelArticle.Description,
			&modelArticle.Body, &modelArticle.CreatedAt, &modelArticle.UpdatedAt, &modelArticle.AuthorID)
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

func (c *Core) FavouriteArticleByArticleId(articleIdList []int64, user *auth.User) (map[int64]bool, error) {
	result := map[int64]bool{}
	for _, articleId := range articleIdList {
		result[articleId] = false
	}
	if user == nil {
		return result, nil
	}

	placeholders, args := stringutils.INCluse(articleIdList)
	selectSQL := fmt.Sprintf(`
		SELECT article_id FROM favourite_articles WHERE user_id = $1 and article_id in (%s)
	`, strings.Join(placeholders, ","))

	queryResult, err := database.ExecuteQuery(c.sqlTemplate, selectSQL, func(rows *sql.Rows) (int64, error) {
		var articleId int64
		if err := rows.Scan(&articleId); err != nil {
			return 0, xerrors.New(err)
		}
		return articleId, nil
	}, args...)

	if err != nil {
		return nil, xerrors.New(err)
	}

	for _, articleId := range queryResult {
		result[articleId] = true
	}

	return result, nil
}

func (c *Core) FavouriteCountByArticleId(articleIdList []int64) (map[int64]int, error) {
	result := map[int64]int{}
	for _, articleId := range articleIdList {
		result[articleId] = 0
	}

	placeholders, args := stringutils.INCluse(articleIdList)
	selectSQL := fmt.Sprintf(`
		SELECT article_id FROM favourite_articles WHERE user_id = $1 and article_id in (%s)
	`, strings.Join(placeholders, ","))

	queryResult, err := database.ExecuteQuery(c.sqlTemplate, selectSQL, func(rows *sql.Rows) (int64, error) {
		var articleId int64
		if err := rows.Scan(&articleId); err != nil {
			return 0, xerrors.New(err)
		}
		return articleId, nil
	}, args...)

	if err != nil {
		return nil, xerrors.New(err)
	}

	for _, articleId := range queryResult {
		result[articleId] = true
	}

	return result, nil
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

func (c *Core) GetArticles(filter filter.Filter, tag, authorUserName, favoritedBy string) ([]*models.Article, error) {
	var favoritedById *int64
	if strings.TrimSpace(favoritedBy) != "" {
		user, err := c.GetUserByUsername(favoritedBy)
		if err == nil {
			favoritedById = &user.ID
		}
	}

	selectSQL := `
		SELECT a.id,a.slug,a.title,a.description,a.body,a.created_at,a.updated_at,a.author_id 
		FROM articles AS a 
		    LEFT JOIN articles_tags at ON a.id = at.article_id 
		    LEFT JOIN tags t ON at.tag_id = t.id 
		    LEFT JOIN favourite_articles AS fa ON a.id = fa.article_id 
		    LEFT JOIN users AS u ON a.author_id = u.id     
	`

	whereClause := []string{}
	args := []any{}
	argId := 1

	if tag != "" {
		whereClause = append(whereClause, "t.name = $"+string(rune(argId)))
		args = append(args, tag)
		argId++
	}

	if authorUserName != "" {
		whereClause = append(whereClause, "u.username = $"+string(rune(argId)))
		args = append(args, authorUserName)
		argId++
	}

	if favoritedById != nil {
		whereClause = append(whereClause, "fa.user_id = $"+string(rune(argId)))
		args = append(args, *favoritedById)
		argId++
	}

	if len(whereClause) > 0 {
		selectSQL += " WHERE " + strings.Join(whereClause, " AND ")
	}

	// add limit and offset
	selectSQL += "ORDER BY a.created_at DESC LIMIT $" + string(rune(argId)) + " OFFSET $" + string(rune(argId+1))
	args = append(args, filter.Limit, filter.Offset)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := c.db.QueryContext(ctx, selectSQL, args...)
	if err != nil {
		return nil, xerrors.New(err)
	}
	defer rows.Close()
	var result []*models.Article

	for rows.Next() {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title,
			&article.Description, &article.Body, &article.CreatedAt, &article.UpdatedAt, &article.AuthorID); err != nil {
			return nil, xerrors.New(err)
		}
		result = append(result, &article)
	}
	if err := rows.Err(); err != nil {
		return nil, xerrors.New(err)
	}

	return result, nil
}
