package databaseutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type txKey struct {
}

var ErrNoRowsFound = errors.New("no rows found")

// SQLExecutor defines the common methods implemented by both *sql.DB and *sql.Tx.
// This allows repository methods to work seamlessly with either a direct DB connection
// or an active transaction.
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// SQLSession manages transactions for a database pool.
// It can either wrap a *sql.DB (for non-transactional operations or to begin new txs)
// or a *sql.Tx (when an active transaction is in progress).
type SQLSession struct {
	db      *sql.DB         // The original database pool
	tx      *sql.Tx         // The active transaction, if any
	ctx     context.Context // Context associated with this session instance
	timeout time.Duration
}

// NewSession creates a new SQLSession wrapping the provided *sql.DB.
func NewSession(db *sql.DB, timeout time.Duration) *SQLSession {
	return &SQLSession{
		db:      db,
		timeout: timeout,
	}
}

// BeginTx starts a new transaction from the DB pool.
// It returns a new SQLSession wrapping this transaction.
func (s *SQLSession) BeginTx(ctx context.Context, opts *sql.TxOptions) (*SQLSession, error) {
	tx, err := s.db.BeginTx(ctx, opts) // Begin a transaction from the pool
	if err != nil {
		return nil, fmt.Errorf("session: failed to begin transaction: %w", err)
	}

	// Return a new session instance that holds this transaction and a context
	// containing the transaction.
	return &SQLSession{
		db:  s.db,
		tx:  tx,
		ctx: nil,
	}, nil
}

// DoTransactionally executes fn within a new transaction. It commits when fn
// succeeds and rolls back when fn returns an error or panics.
func (s *SQLSession) DoTransactionally[T any](ctx context.Context, fn func(tx *sql.Tx) (T, error)) (result T, err error) {
	session, err := s.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = session.Rollback()
			panic(p)
		}

		if err != nil {
			var zero T
			result = zero
			if rollbackErr := session.Rollback(); rollbackErr != nil {
				err = errors.Join(err, fmt.Errorf("session: failed to rollback transaction: %w", rollbackErr))
			}
			return
		}

		if commitErr := session.Commit(); commitErr != nil {
			var zero T
			result = zero
			err = fmt.Errorf("session: failed to commit transaction: %w", commitErr)
		}
	}()

	result, err = fn(session.tx)
	return
}

func (s *SQLSession) Rollback() error {
	if s.tx == nil {
		return fmt.Errorf("session: no active transaction to rollback")
	}
	return s.tx.Rollback()
}

// Commit commits the transaction held by this session.
func (s *SQLSession) Commit() error {
	if s.tx == nil {
		return fmt.Errorf("session: no active transaction to commit")
	}
	return s.tx.Commit()
}

// Context returns the context associated with this session instance.
func (s *SQLSession) Context() context.Context {
	return s.ctx
}

func (s *SQLSession) GetSQLExecutor(tx *sql.Tx) SQLExecutor {
	if tx == nil {
		// If no transaction in context, use the fallback *sql.DB.
		// Operations on *sql.DB auto-commit single statements.
		return s.db
	}

	// If a transaction (*sql.Tx) is found in the context, return it.
	return tx
}

func (s *SQLSession) ExecuteQuery[T any](ctx context.Context, tx *sql.Tx, sql string, extractor func(rows *sql.Rows) (T, error), args ...any) ([]T, error) {
	ctx, cancel, err := contextTimeoutAware(s.timeout, ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	executor := s.GetSQLExecutor(tx)
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

func (s *SQLSession) ExecuteSingleQuery[T any](ctx context.Context, tx *sql.Tx, query string, extractor func(row *sql.Row) (T, error), args ...any) (T, error) {
	ctx, cancel, err := contextTimeoutAware(s.timeout, ctx)
	if err != nil {
		var zeroT T
		return zeroT, err
	}
	defer cancel()

	executor := s.GetSQLExecutor(tx)
	row := executor.QueryRowContext(ctx, query, args...)

	result, err := extractor(row)
	if errors.Is(err, sql.ErrNoRows) {
		var zeroT T
		return zeroT, ErrNoRowsFound
	}
	return result, err
}

func (s *SQLSession) ExecuteNonQuery(ctx context.Context, tx *sql.Tx, sql string, args ...any) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel, err := contextTimeoutAware(s.timeout, ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	executor := s.GetSQLExecutor(tx)
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
