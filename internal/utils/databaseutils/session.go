package databaseutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ctxKey identifies the transaction stored in a context by SQLSession.
type ctxKey struct{}


var ErrNoRowsFound = errors.New("no rows found")

// SQLExecutor defines the common methods implemented by both *sql.DB and *sql.Tx.
// This allows repository methods to work seamlessly with either a direct DB connection
// or an active transaction.
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// SQLSession manages queries and transactions for a database pool.
type SQLSession struct {
	db      *sql.DB // The original database pool
	timeout time.Duration
}

// NewSession creates a new SQLSession wrapping the provided *sql.DB.
func NewSession(db *sql.DB, timeout time.Duration) *SQLSession {
	return &SQLSession{
		db:      db,
		timeout: timeout,
	}
}

// BeginTx returns a derived context containing the transaction and its handle.
// The caller owns committing or rolling back the transaction. A context that
// already contains a SQLSession transaction cannot start another one.
func (s *SQLSession) BeginTx(ctx context.Context, opts *sql.TxOptions) (context.Context, *sql.Tx, error) {

	if _, ok := ctx.Value(ctxKey{}).(*sql.Tx); ok {
		return ctx, nil, fmt.Errorf("session: nested transactions are not supported")
	}

	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return ctx, nil, fmt.Errorf("session: failed to begin transaction: %w", err)
	}

	return context.WithValue(ctx, ctxKey{}, tx), tx, nil
}

// DoTransactionally executes fn within a new transaction. It commits when fn
// succeeds and rolls back when fn returns an error or panics. All operations in
// fn must use its supplied context, which must not be used after fn returns.
func (s *SQLSession) DoTransactionally[T any](ctx context.Context, fn func(txCtx context.Context) (T, error)) (result T, err error) {
	txCtx, tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			var zero T
			result = zero
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				err = errors.Join(err, fmt.Errorf("session: failed to rollback transaction: %w", rollbackErr))
			}
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			var zero T
			result = zero
			err = fmt.Errorf("session: failed to commit transaction: %w", commitErr)
		}
	}()

	result, err = fn(txCtx)
	return
}

// GetSQLExecutor uses the transaction stored in ctx, or this session's pool
// when the context contains no transaction.
func (s *SQLSession) GetSQLExecutor(ctx context.Context) SQLExecutor {
	if tx, ok := ctx.Value(ctxKey{}).(*sql.Tx); ok {
		return tx
	}
	return s.db
}

func (s *SQLSession) ExecuteQuery[T any](ctx context.Context, sql string, extractor func(rows *sql.Rows) (T, error), args ...any) ([]T, error) {
	ctx, cancel, err := contextTimeoutAware(s.timeout, ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	executor := s.GetSQLExecutor(ctx)
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

func (s *SQLSession) ExecuteSingleQuery[T any](ctx context.Context, query string, extractor func(row *sql.Row) (T, error), args ...any) (T, error) {
	ctx, cancel, err := contextTimeoutAware(s.timeout, ctx)
	if err != nil {
		var zeroT T
		return zeroT, err
	}
	defer cancel()

	executor := s.GetSQLExecutor(ctx)
	row := executor.QueryRowContext(ctx, query, args...)

	result, err := extractor(row)
	if errors.Is(err, sql.ErrNoRows) {
		var zeroT T
		return zeroT, ErrNoRowsFound
	}
	return result, err
}

func (s *SQLSession) ExecuteNonQuery(ctx context.Context, sql string, args ...any) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel, err := contextTimeoutAware(s.timeout, ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	executor := s.GetSQLExecutor(ctx)
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
