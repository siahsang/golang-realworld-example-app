package databaseutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type txKey struct {
}

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
	db  *sql.DB         // The original database pool
	tx  *sql.Tx         // The active transaction, if any
	ctx context.Context // Context associated with this session instance
}

// NewSession creates a new SQLSession wrapping the provided *sql.DB.
func NewSession(db *sql.DB) *SQLSession {
	return &SQLSession{
		db: db,
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
	txCtx := context.WithValue(ctx, txKey{}, tx)
	return &SQLSession{
		db:  s.db,
		tx:  tx,
		ctx: txCtx,
	}, nil
}

// DoTransactionally executes fn within a new transaction. It commits when fn
// succeeds and rolls back when fn returns an error or panics.
func (s *SQLSession) DoTransactionally[T any](ctx context.Context, fn func(txCtx context.Context) (T, error)) (result T, err error) {
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

	result, err = fn(session.Context())
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

// GetExecutor returns the current transaction (if active) or the underlying DB pool.
// This is the function called by repositories.
func (s *SQLSession) GetExecutor() SQLExecutor {
	if s.tx != nil {
		return s.tx // Return the active *sql.Tx if present
	}
	return s.db // If no active transaction, return the *sql.DB pool
}

// GetSQLExecutor is a public helper function for repositories to retrieve the
// correct database handle from the context.
// If a transaction (*sql.Tx) is present in the context, it returns that transaction.
// Otherwise, it returns the fallback *sql.DB connection.
func GetSQLExecutor(ctx context.Context, fallbackDB *sql.DB) SQLExecutor {
	// Check if a transaction is stored in the context by WithTransaction or Begin
	dbExecutor := ctx.Value(txKey{})

	if dbExecutor == nil {
		// If no transaction in context, use the fallback *sql.DB.
		// Operations on *sql.DB auto-commit single statements.
		return fallbackDB
	}

	// If a transaction (*sql.Tx) is found in the context, return it.
	tx, ok := dbExecutor.(*sql.Tx)
	if !ok {
		// This indicates a type mismatch if something other than *sql.Tx
		// was stored with txKey. Panic as it's a critical error.
		panic(fmt.Sprintf("session: value in context for txKey is not a *sql.Tx, but %T", dbExecutor))
	}
	return tx
}
