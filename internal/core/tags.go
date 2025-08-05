package core

import (
	"context"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/models"
	"log/slog"
	"strings"
	"time"
)

func (c *Core) CreateTag(tags []*models.Tag) ([]*models.Tag, error) {

	if len(tags) == 0 {
		return nil, xerrors.New("No tags provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, xerrors.New(err)
	}

	defer tx.Rollback()

	// The SQL statement will look like: INSERT INTO tags (name) VALUES ($1), ($2), ...
	valueString := make([]string, 0, len(tags))
	valueArgs := make([]any, 0, len(tags)*2)

	for i, tag := range tags {
		valueString = append(valueString, fmt.Sprintf("($%d)", i+1))
		valueArgs = append(valueArgs, tag.Name)
	}

	// Join the value strings to create the full VALUES clause.
	// e.g., "($1),($2),($3)"
	valueCluses := strings.Join(valueString, ", ")

	// Construct the full SQL statement.

	insertSQL := fmt.Sprintf(`
			INSERT INTO tags (name)
		  	VALUES %s	
		  	ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		  	RETURNING id, name
`, valueCluses)

	rows, err := tx.QueryContext(ctx, insertSQL, valueArgs...)
	if err != nil {
		return nil, xerrors.New(err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			// Log the close error, but it might not be the primary error.
			c.log.Error("failed to close rows", slog.String("error", err.Error()))
		}
	}()

	// Use a map to efficiently store and retrieve the returned tags by name.
	// This helps correctly match the returned IDs to the original input tags,
	// especially if the order of results isn't guaranteed to match the input.
	returnTagsMap := make(map[string]*models.Tag)
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, xerrors.Newf("failed to scan returned tag: %w", err)
		}
		returnTagsMap[name] = &models.Tag{
			ID:   id,
			Name: name,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, xerrors.Newf("error iterating over returned tags: %w", err)
	}

	resultTags := make([]*models.Tag, 0, len(tags))
	for _, tag := range tags {
		if existingTag, exists := returnTagsMap[tag.Name]; exists {
			tag.ID = existingTag.ID // Assign the ID from the database tag
			resultTags = append(resultTags, existingTag)
		} else {
			return nil, xerrors.Newf("tag %s not found in database", tag.Name)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, xerrors.Newf("failed to commit transaction: %w", err)
	}

	return resultTags, nil

}
