package core

import (
	"context"
	"database/sql"

	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/models"
)

func (c *Core) CreateComment(context context.Context, comment *models.Comment) (*models.Comment, error) {
	insertSQL := `
		INSERT INTO comments (body,created_at,updated_at,author_id,article_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id,body,created_at,updated_at,author_id,article_id
	`

	newComment, err := c.session.ExecuteSingleQuery(context, insertSQL, func(row *sql.Row) (*models.Comment, error) {
		var comment models.Comment
		if err := row.Scan(&comment.ID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt, &comment.AuthorID, &comment.ArticleID); err != nil {
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
	comments, err := c.session.ExecuteQuery(context, query, func(rows *sql.Rows) (*models.Comment, error) {
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

func (c *Core) GetCommentByIdAndSlug(ctx context.Context, commentID int64, slug string) (*models.Comment, error) {
	query := `
		SELECT c.id,c.body,c.created_at,c.updated_at,c.author_id,c.article_id
		FROM comments c
		JOIN articles a ON a.id = c.article_id
		WHERE c.id = $1 AND a.slug = $2
	`
	comment, err := c.session.ExecuteSingleQuery(ctx, query, func(row *sql.Row) (*models.Comment, error) {
		var comment models.Comment
		if err := row.Scan(&comment.ID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt, &comment.AuthorID, &comment.ArticleID); err != nil {
			return nil, xerrors.New(err)
		}
		return &comment, nil
	}, commentID, slug)
	if err != nil {
		return nil, xerrors.New(err)
	}
	return comment, nil
}

func (c *Core) DeleteCommentById(ctx context.Context, commentId int64) (int64, error) {
	deleteSQL := `
		DELETE FROM comments
		WHERE id =$1
	`
	rowAffected, err := c.session.ExecuteNonQuery(ctx, deleteSQL, commentId)
	if err != nil {
		return -1, xerrors.New(err)
	}
	return rowAffected, nil
}
