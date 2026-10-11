package databaseutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var ErrNoRowsFound = fmt.Errorf("no Rows Found")

type DBStore struct {
	DB      *sql.DB
	Timeout time.Duration
}

func NewDBStore(db *sql.DB, timeout time.Duration) *DBStore {
	return &DBStore{
		DB:      db,
		Timeout: timeout,
	}
}

func ExecuteQuery[T any](dbStore *DBStore, ctx context.Context, sql string, extractor func(rows *sql.Rows) (T, error), args ...any) ([]T, error) {
	ctx, cancel, err := contextTimeoutAware(dbStore.Timeout, ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	executor := GetSQLExecutor(ctx, dbStore.DB)
	rows, err := executor.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results = make([]T, 0)
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

func ExecuteSingleQuery[T any](dbStore *DBStore, ctx context.Context, query string, extractor func(row *sql.Row) (T, error), args ...any) (T, error) {
	ctx, cancel, err := contextTimeoutAware(dbStore.Timeout, ctx)
	if err != nil {
		var zeroT T
		return zeroT, err
	}
	defer cancel()

	executor := GetSQLExecutor(ctx, dbStore.DB)
	row := executor.QueryRowContext(ctx, query, args...)

	return extractor(row)
}

func ExecuteNonQuery(dbStore *DBStore, ctx context.Context, sql string, args ...any) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel, err := contextTimeoutAware(dbStore.Timeout, ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	executor := GetSQLExecutor(ctx, dbStore.DB)
	result, err := executor.ExecContext(ctx, sql, args...)
	if err != nil {
		return -1, err
	}
	affected, err := result.RowsAffected()

	return affected, err
}

// contextTimeoutAware returns an error if ctx is already done. For a positive
// duration, it derives a timeout context; the parent's earlier deadline still
// takes precedence. For a non-positive duration, it returns ctx unchanged with
// a no-op cancel function so callers can always defer cancel safely.
func contextTimeoutAware(duration time.Duration, ctx context.Context) (context.Context, context.CancelFunc, error) {
	// Avoid starting database work when the caller's context is already done.
	if err := ctx.Err(); err != nil {
		return nil, nil, ctx.Err()
	}

	// A non-positive duration disables the additional query timeout.
	if duration <= 0 {
		return ctx, func() {}, nil
	}

	childCtx, cancel := context.WithTimeout(ctx, duration)

	return childCtx, cancel, nil
}
