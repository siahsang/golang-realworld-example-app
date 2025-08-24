package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/utils/collectionutils"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/models"
	"strings"
	"time"
)

func (c *Core) CreateTag(context context.Context, tags []*models.Tag) ([]*models.Tag, error) {

	if len(tags) == 0 {
		return nil, xerrors.New("No tags provided")
	}

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

	tagList, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, insertSQL, func(rows *sql.Rows) (*models.Tag, error) {
		tag := &models.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, xerrors.Newf("failed to scan returned tag: %w", err)
		}

		return tag, nil
	}, valueArgs...)

	if err != nil {
		return nil, xerrors.Newf("failed to insert tags: %w", err)
	}

	// Use a map to efficiently store and retrieve the returned tags by name.
	// This helps correctly match the returned IDs to the original input tags,
	// especially if the order of results isn't guaranteed to match the input.
	returnTagsMap := collectionutils.Associate(tagList, func(tag *models.Tag) (string, *models.Tag) {
		return tag.Name, tag
	})

	resultTags := make([]*models.Tag, 0, len(tags))
	for _, tag := range tags {
		if existingTag, exists := returnTagsMap[tag.Name]; exists {
			tag.ID = existingTag.ID // Assign the ID from the database tag
			resultTags = append(resultTags, existingTag)
		} else {
			return nil, xerrors.Newf("tag %s not found in database", tag.Name)
		}
	}

	return resultTags, nil

}

func (c *Core) GetTagsByArticleId(articleIdList []int64) (map[int64][]models.Tag, error) {
	if len(articleIdList) == 0 {
		return make(map[int64][]models.Tag), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Build the placeholders for the IN clause
	placeholders := make([]string, len(articleIdList))
	args := make([]any, len(articleIdList))
	for i, id := range articleIdList {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT at.article_id, t.id, t.name
		FROM articles_tags at
		JOIN tags t ON at.tag_id = t.id
		WHERE at.article_id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, xerrors.Newf("failed to query tags by article ids: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]models.Tag)
	for rows.Next() {
		var articleID, tagID int64
		var tagName string
		if err := rows.Scan(&articleID, &tagID, &tagName); err != nil {
			return nil, xerrors.Newf("failed to scan row: %w", err)
		}
		result[articleID] = append(result[articleID], models.Tag{
			ID:   tagID,
			Name: tagName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, xerrors.Newf("row iteration error: %w", err)
	}

	return result, nil
}
