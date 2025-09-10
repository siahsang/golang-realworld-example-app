package core

import (
	"context"
	"database/sql"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/models"
)

func (c *Core) CreateComment(context context.Context, comment *models.Comment) (*models.Comment, error) {
	insertSQL := `
		INSERT INTO comments (body,created_at,updated_at,author_id,article_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id,body,created_at,updated_at,author_id,article_id
	`

	newComment, err := databaseutils.ExecuteSingleQuery(c.sqlTemplate, context, insertSQL, func(rows *sql.Rows) (*models.Comment, error) {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt, &comment.AuthorID, &comment.ArticleID); err != nil {
			return nil, xerrors.New(err)
		}
		return &comment, nil
	}, comment.Body, comment.CreatedAt, comment.UpdatedAt, comment.AuthorID, comment.ArticleID)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return newComment, nil
}
