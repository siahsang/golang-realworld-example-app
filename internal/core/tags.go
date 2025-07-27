package core

import (
	"context"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/models"
	"strings"
	"time"
)

var (
	ErrTagAlreadyExists = xerrors.New("Tag already exists")
)

func (c *Core) CreateTag(tags []*models.Tag) ([]*models.Tag, error) {
	const insertSQL = `	
		INSERT INTO tag (name)
		VALUES ($1)
		RETURNING id, name
	`

	returnTag := &models.Tag{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := c.db.QueryRowContext(ctx, insertSQL, returnTag.ID, returnTag.Name).
		Scan(&returnTag.ID, &returnTag.Name)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint`):
			return nil, xerrors.New(ErrTagAlreadyExists)
		default:
			return nil, xerrors.New(err)
		}
	}

	return returnTag, nil
}
