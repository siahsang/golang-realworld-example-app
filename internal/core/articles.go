package core

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/utils/stringutils"
	"github.com/siahsang/blog/models"
)

var ErrDuplicatedSlug = xerrors.Message("Duplicate slug")
var ErrDuplicatedArticleTag = xerrors.Message("Duplicate article tag")

func (c *Core) CreateArticle(context context.Context, article *models.Article, tagModels []*models.Tag) (*models.Article, error) {

	insertSQL := `
		INSERT INTO articles (slug,title,description,body,created_at,updated_at,author_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id,slug,title,description,body,created_at,updated_at,author_id
	`

	newArticle, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, insertSQL, func(rows *sql.Rows) (*models.Article, error) {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title,
			&article.Description, &article.Body, &article.CreatedAt, &article.UpdatedAt, &article.AuthorID); err != nil {
			return nil, xerrors.New(err)
		}
		return &article, nil
	}, article.Slug, article.Title, article.Description, article.Body, time.Now(), time.Now(), article.AuthorID)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), `duplicate key value violates unique constraint`):
			return nil, xerrors.New(ErrDuplicatedSlug)
		default:
			return nil, xerrors.New(err)
		}
	}
	savedTagList, err := c.CreateTag(context, tagModels)
	if err != nil {
		return nil, xerrors.New(err)
	}

	for _, tag := range savedTagList {
		insertSQL := `
			INSERT INTO articles_tags (article_id, tag_id)
			VALUES ($1, $2)
			RETURNING article_id,tag_id
		`

		type QueryResult struct {
			ArticleID int64
			TagID     int64
		}
		_, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, insertSQL, func(rows *sql.Rows) (*QueryResult, error) {
			qr := &QueryResult{}
			if err := rows.Scan(&qr.ArticleID, &qr.TagID); err != nil {
				return nil, xerrors.New(err)
			}
			return qr, nil
		}, newArticle[0].ID, tag.ID)

		if err != nil {
			switch {
			case strings.Contains(err.Error(), `duplicate key value violates unique constraint`):
				return nil, xerrors.New(ErrDuplicatedArticleTag)
			default:
				return nil, xerrors.New(err)
			}

		}
	}

	return newArticle[0], nil
}

func (c *Core) IsFavouriteArticleByUser(context context.Context, articleId int64, user *auth.User) (bool, error) {
	if user == nil {
		return false, nil
	}

	const selectSQL = `
		SELECT EXISTS( 
			SELECT 1 FROM favourite_articles WHERE user_id = $1 and article_id = $2
		)
	`

	result, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, selectSQL, func(rows *sql.Rows) (bool, error) {
		var isFavourite bool
		if err := rows.Scan(&isFavourite); err != nil {
			return false, xerrors.New(err)
		}
		return isFavourite, nil

	}, user.ID, articleId)

	if err != nil {
		return false, xerrors.New(err)
	}

	return result, nil
}

func (c *Core) FavouriteArticleByArticleId(context context.Context, articleIdList []int64, user *auth.User) (map[int64]bool, error) {
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

	queryResult, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, selectSQL, func(rows *sql.Rows) (int64, error) {
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

func (c *Core) FavouriteCountByArticleId(context context.Context, articleIdList []int64) (map[int64]int64, error) {
	result := map[int64]int64{}
	for _, articleId := range articleIdList {
		result[articleId] = 0
	}

	if len(articleIdList) == 0 {
		return result, nil
	}

	placeholders, args := stringutils.INCluse(articleIdList)
	selectSQL := fmt.Sprintf(`
		SELECT COUNT(*) as count, article_id
		FROM favourite_articles		
		WHERE article_id IN (%s)
		GROUP BY article_id
	`, strings.Join(placeholders, ","))

	type QueryResult struct {
		ArticleId int64
		Count     int64
	}

	queryResultList, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, selectSQL, func(rows *sql.Rows) (*QueryResult, error) {
		queryResult := &QueryResult{}

		if err := rows.Scan(&queryResult.Count, &queryResult.ArticleId); err != nil {
			return nil, xerrors.New(err)
		}
		return queryResult, nil
	}, args...)

	if err != nil {
		return nil, xerrors.New(err)
	}

	for _, q := range queryResultList {
		result[q.ArticleId] = q.Count
	}

	return result, nil
}

func (c *Core) FavouriteArticleCount(context context.Context, articleId int64) (int64, error) {
	const selectSQL = `
		SELECT COUNT(*) FROM favourite_articles WHERE article_id = $1
	`

	result, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, selectSQL, func(rows *sql.Rows) (int64, error) {
		var favouriteArticleCount int64
		if err := rows.Scan(&favouriteArticleCount); err != nil {
			return 0, xerrors.New(err)
		}
		return favouriteArticleCount, nil
	}, articleId)

	if err != nil {
		return 0, xerrors.New(err)
	}

	return result, nil
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

func (c *Core) GetArticles(context context.Context, filter filter.Filter, tag, authorUserName, favoritedBy string) ([]*models.Article, error) {
	var favoritedById *int64
	if strings.TrimSpace(favoritedBy) != "" {
		user, err := c.GetUserByUsername(context, favoritedBy)
		if err == nil {
			favoritedById = &user.ID
		}
	}

	selectSQL := `
		SELECT DISTINCT a.id,a.slug,a.title,a.description,a.body,a.created_at,a.updated_at,a.author_id
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
		whereClause = append(whereClause, " t.name = $"+fmt.Sprintf("%d", argId))
		args = append(args, tag)
		argId++
	}

	if authorUserName != "" {
		whereClause = append(whereClause, " u.username = $"+fmt.Sprintf("%d", argId))
		args = append(args, authorUserName)
		argId++
	}

	if favoritedById != nil {
		whereClause = append(whereClause, " fa.user_id = $"+fmt.Sprintf("%d", argId))
		args = append(args, *favoritedById)
		argId++
	}

	if len(whereClause) > 0 {
		selectSQL += " WHERE " + strings.Join(whereClause, " AND ")
	}

	// add limit and offset
	selectSQL += " ORDER BY a.created_at DESC LIMIT $" + fmt.Sprintf("%d", argId) + " OFFSET $" + fmt.Sprintf("%d", argId+1)
	args = append(args, filter.Limit, filter.Offset)

	result, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, selectSQL, func(rows *sql.Rows) (*models.Article, error) {
		var article = &models.Article{}
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title,
			&article.Description, &article.Body, &article.CreatedAt, &article.UpdatedAt, &article.AuthorID); err != nil {
			return nil, xerrors.New(err)
		}
		return article, nil
	}, args...)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return result, nil
}

func (c *Core) UpdateArticle(context context.Context, article *models.Article) (*models.Article, error) {
	query := `
		UPDATE articles
		SET title = $1, description = $2, body = $3, updated_at = $4
		WHERE id = $5
		RETURNING id,slug,title,description,body,created_at,updated_at,author_id
	`
	args := []any{article.Title, article.Description, article.Body, time.Now(), article.ID}
	returningArticle, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*models.Article, error) {
		var article = &models.Article{}
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title,
			&article.Description, &article.Body, &article.CreatedAt, &article.UpdatedAt, &article.AuthorID); err != nil {
			return nil, xerrors.New(err)
		}
		return article, nil
	}, args...)

	if err != nil {
		return nil, xerrors.New(err)
	}
	return returningArticle, nil

}

func (c *Core) GetArticleBySlug(context context.Context, slug string) (*models.Article, error) {
	selectSQL := `
		SELECT a.id,a.slug,a.title,a.description,a.body,a.created_at,a.updated_at,a.author_id
		FROM articles AS a 
		WHERE a.slug = $1
	`

	result, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, selectSQL, func(rows *sql.Rows) (*models.Article, error) {
		var article = &models.Article{}
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title,
			&article.Description, &article.Body, &article.CreatedAt, &article.UpdatedAt, &article.AuthorID); err != nil {
			return nil, xerrors.New(err)
		}
		return article, nil
	}, slug)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return result, nil
}

func (c *Core) FavoriteArticle(context context.Context, slug string, user *auth.User) (*models.Article, error) {
	article, err := c.GetArticleBySlug(context, slug)
	if err != nil {
		return nil, xerrors.New(err)
	}

	const updateSQL = `
		INSERT INTO favourite_articles (user_id, article_id)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT favourite_articles_pkey DO NOTHING
		RETURNING user_id,article_id
	`

	_, err = databaseutils.ExecuteNonQuery(c.sqlTemplate, context, updateSQL, user.ID, article.ID)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return article, nil
}
