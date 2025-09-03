package databaseutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var ErrNoRowsFound = fmt.Errorf("no Rows Found")

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
	ctx, cancel, err := contextTimeOutAware(sqlTemplate.Timeout, ctx, cancel)
	if err != nil {
		return nil, err
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

func ExecuteSingleQuery[T any](sqlTemplate *SQLTemplate, ctx context.Context, sql string, extractor func(rows *sql.Rows) (T, error), args ...any) (T, error) {
	query, err := ExecuteQuery(sqlTemplate, ctx, sql, extractor, args...)
	if err != nil {
		var empty T
		return empty, err
	} else {
		if len(query) == 0 {
			var empty T
			return empty, ErrNoRowsFound
		}
		return query[0], nil
	}
}

func ExecuteDeleteQuery(sqlTemplate *SQLTemplate, ctx context.Context, sql string, args ...any) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel, err := contextTimeOutAware(sqlTemplate.Timeout, ctx, cancel)
	if err != nil {
		return 0, err
	}
	defer cancel()

	executor := GetSQLExecutor(ctx, sqlTemplate.DB)
	result, err := executor.ExecContext(ctx, sql, args)
	if err != nil {
		return -1, err
	}
	affected, err := result.RowsAffected()

	return affected, err
}

func contextTimeOutAware(duration time.Duration, ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc, error) {
	if duration > 0 {
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return nil, nil, ctx.Err()
			}
			if remaining > duration {
				ctx, cancel = context.WithTimeout(ctx, duration)
			}
		} else {
			ctx, cancel = context.WithTimeout(ctx, duration)
		}
	}
	return ctx, cancel, nil
}
