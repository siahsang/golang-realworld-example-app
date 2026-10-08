package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/internal/filter"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/utils/functional"
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

	var savedTagList []*models.Tag
	if len(tagModels) > 0 {
		savedTagList, err = c.CreateTag(context, tagModels)
		if err != nil {
			return nil, xerrors.New(err)
		}
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

func (c *Core) FavouriteArticleByUser(context context.Context, articleIdList []int64, user *auth.User) (map[int64]bool, error) {
	result := map[int64]bool{}
	for _, articleId := range articleIdList {
		result[articleId] = false
	}
	if user == nil {
		return result, nil
	}

	if len(articleIdList) == 0 {
		return result, nil
	}

	placeholders, articleArgs := stringutils.INClause(articleIdList, 2)
	args := make([]any, 0, len(articleArgs)+1)
	args = append(args, user.ID)
	args = append(args, articleArgs...)

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

	placeholders, args := stringutils.INCluseOld(articleIdList)
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

func (c *Core) CreateSlug(title string, ctx context.Context) (string, error) {
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

	for {
		exist, err := c.isSlugExist(slug, ctx)
		if err != nil {
			return "", err
		}

		if !exist {
			break
		} else {
			c.log.Info("slug already exist tring again", "slug", slug)
			slug = fmt.Sprintf("%s-%s", slug, time.Now().Format("2006-01-02")[:8])
		}
	}

	return slug, nil
}

func (c *Core) GetArticles(context context.Context, filter filter.Filter, tag, authorUserName, favoriteBy string) ([]*models.Article, int64, error) {
	var favoritedById *int64
	if strings.TrimSpace(favoriteBy) != "" {
		user, err := c.GetUserByUsername(context, favoriteBy)
		if err == nil {
			favoritedById = &user.ID
		} else {
			if errors.Is(err, NoRecordFound) {
				return []*models.Article{}, 0, nil
			}
			return nil, 0, err
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

	countSQL := `
		SELECT COUNT(DISTINCT a.id)
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
		countSQL += " WHERE " + strings.Join(whereClause, " AND ")
	}

	totalCount, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, countSQL, func(rows *sql.Rows) (int64, error) {
		var totalCount int64
		if err := rows.Scan(&totalCount); err != nil {
			return -1, xerrors.New(err)
		}
		return totalCount, nil
	}, args...)

	if err != nil {
		return nil, -1, xerrors.New(err)
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
		return nil, -1, xerrors.New(err)
	}

	return result, totalCount, nil
}

func (c *Core) FeedArticle(ctx context.Context, username string, limit, offset int64) ([]*models.Article, int64, error) {
	followingUserList, err := c.GetFollowingUserList(ctx, username)
	if err != nil {
		return nil, -1, xerrors.New(err)
	}

	userListId := functional.Map(followingUserList, func(user *auth.User) int64 {
		return user.ID
	})

	if len(userListId) == 0 {
		return []*models.Article{}, 0, nil
	}

	placeholders, args := stringutils.INClause(userListId, 1)
	inClause := strings.Join(placeholders, ",")

	selectSQL := fmt.Sprintf(`
	SELECT DISTINCT
		a.id, a.slug, a.title, a.description, a.body,
		a.created_at, a.updated_at, a.author_id
	FROM articles AS a
	WHERE a.author_id IN (%s)
`, inClause)

	countSQL := fmt.Sprintf(`
	SELECT COUNT(DISTINCT a.id)
	FROM articles AS a
	WHERE a.author_id IN (%s)
`, inClause)

	totalCount, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, ctx, countSQL, func(rows *sql.Rows) (int64, error) {
		var totalCount int64
		if err := rows.Scan(&totalCount); err != nil {
			return -1, xerrors.New(err)
		}
		return totalCount, nil
	}, args...)

	if err != nil {
		return nil, -1, xerrors.New(err)
	}

	nextArg := len(args) + 1
	selectSQL += fmt.Sprintf(" ORDER BY a.created_at DESC, a.id DESC LIMIT $%d OFFSET $%d", nextArg, nextArg+1)
	args = append(args, limit, offset)

	result, err := databaseutils.ExecuteQuery(c.sqlTemplate, ctx, selectSQL, func(rows *sql.Rows) (*models.Article, error) {
		var article = &models.Article{}
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title,
			&article.Description, &article.Body, &article.CreatedAt, &article.UpdatedAt, &article.AuthorID); err != nil {
			return nil, xerrors.New(err)
		}
		return article, nil
	}, args...)

	if err != nil {
		return nil, -1, xerrors.New(err)
	}

	return result, totalCount, nil

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

func (c *Core) isSlugExist(slug string, ctx context.Context) (bool, error) {

	const articleCountBySlug = `SELECT COUNT(ID) FROM articles WHERE slug=$1`

	exist, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, ctx, articleCountBySlug, func(rows *sql.Rows) (bool, error) {
		var count int
		if err := rows.Scan(&count); err != nil {
			return false, xerrors.New(err)
		}
		return count > 0, nil
	}, slug)

	if err != nil {
		return false, err
	}

	return exist, nil
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

func (c *Core) UnFavoriteArticle(context context.Context, slug string, user *auth.User) (*models.Article, error) {
	article, err := c.GetArticleBySlug(context, slug)
	if err != nil {
		return nil, xerrors.New(err)
	}

	const deleteSQL = `
		DELETE FROM favourite_articles
		WHERE user_id = $1 AND article_id = $2
		RETURNING user_id, article_id
	`

	_, err = databaseutils.ExecuteNonQuery(c.sqlTemplate, context, deleteSQL, user.ID, article.ID)

	if err != nil {
		return nil, xerrors.New(err)
	}
	c.log.Info("user unfavorited article", "id", user.ID, "article_id", article.ID)
	return article, nil
}
