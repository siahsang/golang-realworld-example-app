package databaseutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type txKey struct {
}

// SQLExecutor defines the common methods implemented by both *sql.DB and *sql.Tx.
// This allows repository methods to work seamlessly with either a direct DB connection
// or an active transaction.
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Session interface defines the contract for transaction management.
type Session interface {
	// Begin starts a new database transaction and returns a new Session
	// instance that represents this transaction.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Session, error)

	// DoTransactionally executes a function 'f' within a new transaction.
	// The context passed to 'f' will contain the transaction.
	// The transaction is committed if 'f' returns nil, otherwise it's rolled back.
	DoTransactionally(ctx context.Context, fn func(txCtx context.Context) error) error

	// Rollback rolls back the current transaction.
	Rollback() error

	// Commit commits the current transaction.
	Commit() error

	// Context returns the context associated with this Session.
	// If it's a transactional session, this context contains the *sql.Tx.
	Context() context.Context

	// GetExecutor provides the underlying *sql.Tx (if active) or *sql.DB (for standalone operations).
	// This is used by repositories to get the database connection/transaction.
	GetExecutor() SQLExecutor
}

// sqlSession implements the Session interface.
// It can either wrap a *sql.DB (for non-transactional operations or to begin new txs)
// or a *sql.Tx (when an active transaction is in progress).
type sqlSession struct {
	db  *sql.DB         // The original database pool
	tx  *sql.Tx         // The active transaction, if any
	ctx context.Context // Context associated with this session instance
}

// NewSession creates a new Session instance wrapping the provided *sql.DB.
func NewSession(db *sql.DB) Session {
	return &sqlSession{
		db: db,
	}
}

// Begin starts a new transaction from the DB pool.
// It returns a *new* sqlSession wrapping this transaction.
func (s *sqlSession) BeginTx(ctx context.Context, opts *sql.TxOptions) (Session, error) {
	tx, err := s.db.BeginTx(ctx, opts) // Begin a transaction from the pool
	if err != nil {
		return nil, fmt.Errorf("session: failed to begin transaction: %w", err)
	}

	// Return a new session instance that holds this transaction and a context
	// containing the transaction.
	txCtx := context.WithValue(ctx, txKey{}, tx)
	return &sqlSession{
		db:  s.db,
		tx:  tx,
		ctx: txCtx,
	}, nil
}

// DoTransactionally executes a function 'f' within a new transaction.
// It handles the begin, commit, and rollback logic.
func (s *sqlSession) DoTransactionally(ctx context.Context, fn func(txCtx context.Context) error) error {
	// 1. Directly begin a new *sql.Tx
	session, err := s.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("session: failed to begin transaction: %w", err)
	}

	// 3. Define the defer function for cleanup and commit/rollback.
	// The named return parameter 'err' is crucial here.
	defer func() {
		if p := recover(); p != nil {
			// A panic occurred, so rollback the transaction.
			_ = session.Rollback() // Ignore rollback error on panic, re-panic.
			panic(p)
		} else if err != nil { // 'err' here is the error returned by 'fn'
			// fn returned an error, so rollback.
			if rollbackErr := session.Rollback(); rollbackErr != nil {
				log.Printf("session: failed to rollback transaction after error: %v (original error: %v)", rollbackErr, err)
				// Optionally, you might want to wrap the original error with the rollback error
				// err = fmt.Errorf("original error: %w; rollback error: %v", err, rollbackErr)
			}
		} else {
			// fn succeeded (err is nil), so commit.
			if commitErr := session.Commit(); commitErr != nil {
				// If commit fails, this becomes the primary error for Transactional.
				err = fmt.Errorf("session: failed to commit transaction: %w", commitErr)
			}
		}
	}()

	// 4. Execute the provided function and assign its result to the named return parameter 'err'
	err = fn(session.Context()) // This sets the 'err' for the defer to act upon

	// 5. No explicit commit/rollback here. The defer handles it.
	// The value of 'err' will be returned as set by 'fn' or by a failed commit in the defer.
	return err
}

func (s *sqlSession) Rollback() error {
	if s.tx == nil {
		return fmt.Errorf("session: no active transaction to rollback")
	}
	return s.tx.Rollback()
}

// Commit commits the transaction held by this session.
func (s *sqlSession) Commit() error {
	if s.tx == nil {
		return fmt.Errorf("session: no active transaction to commit")
	}
	return s.tx.Commit()
}

// Context returns the context associated with this session instance.
func (s *sqlSession) Context() context.Context {
	return s.ctx
}

// GetExecutor returns the current transaction (if active) or the underlying DB pool.
// This is the function called by repositories.
func (s *sqlSession) GetExecutor() SQLExecutor {
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

func DoTransactionally[T any](ctx context.Context, session Session, fn func(txCtx context.Context) (T, error)) (T, error) {
	var zero T
	var result T
	err := session.DoTransactionally(ctx, func(txCtx context.Context) error {
		r, err := fn(txCtx)
		result = r
		return err
	})
	if err != nil {
		return zero, err
	}
	return result, nil
}
