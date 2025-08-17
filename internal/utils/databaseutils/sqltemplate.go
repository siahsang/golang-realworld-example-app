package databaseutils

import (
	"context"
	"database/sql"
	"time"
)

type SQLTemplate struct {
	DB      *sql.DB
	Timeout time.Duration
}

func NewSQLTemplate(db *sql.DB, timeout time.Duration) *SQLTemplate {
	return &SQLTemplate{
		DB:      db,
		Timeout: timeout,
	}
}

func ExecuteQuery[T any](sqlTemplate *SQLTemplate, sql string, extractor func(rows *sql.Rows) (T, error), args ...any) ([]T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sqlTemplate.Timeout)
	defer cancel()
	rows, err := sqlTemplate.DB.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		t, err := extractor(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
