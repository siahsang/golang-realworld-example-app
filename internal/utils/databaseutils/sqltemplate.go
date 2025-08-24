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

func ExecuteQuery[T any](sqlTemplate *SQLTemplate, ctx context.Context, sql string, extractor func(rows *sql.Rows) (T, error), args ...any) ([]T, error) {
	var cancel context.CancelFunc
	if sqlTemplate.Timeout > 0 {
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return nil, ctx.Err()
			}
			if remaining > sqlTemplate.Timeout {
				context.WithTimeout(ctx, sqlTemplate.Timeout)
			}
		} else {
			ctx, cancel = context.WithTimeout(ctx, sqlTemplate.Timeout)
		}
	}

	defer cancel()

	executor := GetSQLExecutor(ctx, sqlTemplate.DB)
	rows, err := executor.QueryContext(ctx, sql, args...)
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
