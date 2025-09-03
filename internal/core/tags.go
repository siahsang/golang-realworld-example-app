package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mdobak/go-xerrors"
	"github.com/siahsang/blog/internal/utils/collectionutils"
	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/internal/utils/functional"
	"github.com/siahsang/blog/models"
	"strings"
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

func (c *Core) GetTagsByArticleId(context context.Context, articleIdList []int64) (map[int64][]models.Tag, error) {
	if len(articleIdList) == 0 {
		return make(map[int64][]models.Tag), nil
	}

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

	type QueryTempResult struct {
		ArticleID int64
		TagID     int64
		TagName   string
	}

	foundArticleList, err := databaseutils.ExecuteQuery(c.sqlTemplate, context, query, func(rows *sql.Rows) (QueryTempResult, error) {
		queryTempResult := QueryTempResult{}
		if err := rows.Scan(&queryTempResult.ArticleID, &queryTempResult.TagID, &queryTempResult.TagName); err != nil {
			return QueryTempResult{}, xerrors.Newf("failed to scan row: %w", err)
		}
		return queryTempResult, nil
	}, args...)

	if err != nil {
		return nil, xerrors.Newf("failed to query tags by article ids: %w", err)
	}

	resultGroupByArticleId := collectionutils.GroupBy(foundArticleList, func(item QueryTempResult) int64 {
		return item.ArticleID
	})

	result := make(map[int64][]models.Tag, len(resultGroupByArticleId))
	for key, value := range resultGroupByArticleId {
		tagList := functional.Map(value, func(item QueryTempResult) models.Tag {
			return models.Tag{
				ID:   item.TagID,
				Name: item.TagName,
			}
		})

		result[key] = tagList
	}

	return result, nil
}
