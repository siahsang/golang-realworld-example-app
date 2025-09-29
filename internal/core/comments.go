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

func (c *Core) GetCommentsBySlug(context context.Context, slug string) ([]*models.Comment, error) {
	bySlug, err := c.GetArticleBySlug(context, slug)
	if err != nil {
		return nil, xerrors.New(err)
	}

	query := `
		SELECT id,body,created_at,updated_at,author_id,article_id
		FROM comments
		WHERE article_id = $1
		ORDER BY created_at DESC
	`
	comments, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (*models.Comment, error) {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt, &comment.AuthorID, &comment.ArticleID); err != nil {
			return nil, xerrors.New(err)
		}
		return &comment, nil
	}, bySlug.ID)

	if err != nil {
		return nil, xerrors.New(err)
	}

	return comments, nil
}

func (c *Core) DeleteCommentById(ctx context.Context, commentId int64) (int64, error) {
	deleteSQL := `
		DELETE FROM comments
		WHERE id =$1
	`
	rowAffected, err := databaseutils.ExecuteNonQuery(c.sqlTemplate, ctx, deleteSQL, commentId)
	if err != nil {
		return -1, xerrors.New(err)
	}
	return rowAffected, nil
}
