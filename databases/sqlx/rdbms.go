//go:generate go tool mockgen -destination=../../.mocking/sqlx_mock/sqlx_mock.go -package=sqlx_mock . RDBMS,ReadQuery,WriterCommand,Tx

package sqlx

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
)

// RDBMS is a high-level interface that combines read and write capabilities
// for interacting with a PostgreSQL database using pgx and Squirrel.
// It abstracts query execution (both raw SQL and Squirrel builders) as well as
// prepared statements and transactions.
type RDBMS interface {
	ReadQuery
	WriterCommand
	queryExecutor
	StmtExecutor
	Close() error
	Ping(ctx context.Context) error
}

// WriterCommand defines write operations (INSERT, UPDATE, DELETE) on the database.
type WriterCommand interface {
	WriterCommandSquirrel
}

// ReadQuery defines read operations (SELECT) on the database.
type ReadQuery interface {
	ReadQuerySquirrel
}

// WriterCommandSquirrel defines write operations using Squirrel SQL builders.
type WriterCommandSquirrel interface {
	// ExecSq executes a write query (INSERT, UPDATE, DELETE) built with Squirrel.
	// If useStmt = true, the query will be executed as a prepared statement.
	ExecSq(ctx context.Context, query squirrel.Sqlizer, useStmt bool) (sql.Result, error)
}

// ReadQuerySquirrel defines read operations using Squirrel SQL builders.
type ReadQuerySquirrel interface {
	// QuerySq executes a SELECT query built with Squirrel and processes multiple rows.
	// The provided callback fn will be called with the result set.
	// If useStmt = true, the query will be executed as a prepared statement.
	QuerySq(ctx context.Context, query squirrel.Sqlizer, useStmt bool, fn func(rows *sql.Rows) error) error

	// QuerySqPagination executes a paginated SELECT query built with Squirrel.
	// countQuery is used to count total rows, and query is the actual SELECT statement.
	// paginationInput controls limit/offset and pagination metadata.
	// The provided callback fn will be called with the paginated rows.
	QuerySqPagination(
		ctx context.Context,
		countQuery squirrel.SelectBuilder,
		query squirrel.SelectBuilder,
		useStmt bool,
		paginationInput primitive.PaginationInput,
		fn func(rows *sql.Rows) error,
	) (primitive.PaginationOutput, error)

	// QueryRowSq executes a SELECT query built with Squirrel and returns a single row.
	// If no rows are found, sql.ErrNoRows is returned.
	// If useStmt = true, the query will be executed as a prepared statement.
	QueryRowSq(ctx context.Context, query squirrel.Sqlizer, useStmt bool) (*sql.Row, error)
}

// queryExecutor defines low-level query execution using raw SQL strings.
// This is useful when Squirrel is not used or when executing ad-hoc queries.
type queryExecutor interface {
	// QueryContext executes a SQL string and returns multiple rows.
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// QueryRowContext executes a SQL string and returns a single row.
	// If no rows are found, sql.ErrNoRows is returned.
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	// ExecContext executes a SQL command (INSERT, UPDATE, DELETE, DDL) and returns the result.
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	// PrepareContext creates a prepared statement for later executions.
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// StmtExecutor defines query execution using prepared statements.
type StmtExecutor interface {
	// QueryStmtContext executes a prepared statement that returns multiple rows.
	QueryStmtContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// QueryRowStmtContext executes a prepared statement that returns a single row.
	QueryRowStmtContext(ctx context.Context, query string, args ...any) (*sql.Row, error)

	// ExecStmtContext executes a prepared statement for write operations.
	ExecStmtContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Tx defines transactional operations on the database.
// Provides methods to execute blocks of code within a transaction context.
type Tx interface {
	// DoTxContext executes the given function within a database transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, the transaction is committed.
	// The context and sql.TxOptions can be used to control isolation level and timeout.
	DoTxContext(ctx context.Context, opt *sql.TxOptions, fn func(ctx context.Context, tx RDBMS) error) error
}
