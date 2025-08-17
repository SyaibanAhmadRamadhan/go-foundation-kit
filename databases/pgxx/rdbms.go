//go:generate go tool mockgen -destination=../../.mocking/pgxx_mock/pgxx_mock.go -package=pgxx_mock . RDBMS,ReadQuery,WriterCommand,Tx
package pgxx

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// RDBMS is a high-level interface for interacting with PostgreSQL using pgx and Squirrel.
// It combines read and write capabilities with both raw SQL execution and Squirrel builders.
type RDBMS interface {
	ReadQuery
	WriterCommand
	queryExecutor
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
	// Returns a pgconn.CommandTag which contains command metadata, such as number of rows affected.
	ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error)
}

// ReadQuerySquirrel defines read operations using Squirrel SQL builders.
type ReadQuerySquirrel interface {
	// QuerySq executes a SELECT query built with Squirrel and invokes fn with the result rows.
	// The rows are automatically closed after fn returns.
	QuerySq(ctx context.Context, query squirrel.Sqlizer, fn func(rows pgx.Rows) error) error

	// QuerySqPagination executes a paginated SELECT query using Squirrel.
	//   - countQuery is used to compute the total number of rows.
	//   - query is the SELECT statement with LIMIT/OFFSET applied.
	//   - paginationInput defines page number and page size.
	// The fn callback is invoked with the result rows, which are closed automatically after fn returns.
	// Returns pagination metadata and any error encountered.
	QuerySqPagination(
		ctx context.Context,
		countQuery, query squirrel.SelectBuilder,
		paginationInput primitive.PaginationInput,
		fn func(rows pgx.Rows) error,
	) (primitive.PaginationOutput, error)

	// QueryRowSq executes a SELECT query built with Squirrel and returns a single row.
	// Errors are typically reported when Scan is called on the returned pgx.Row.
	QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error)
}

// queryExecutor defines low-level query execution using raw SQL strings.
// Useful when Squirrel is not used or for executing ad-hoc queries.
type queryExecutor interface {
	// Query executes a SQL string and returns multiple rows.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow executes a SQL string and returns a single row.
	// Errors are deferred until Scan is called on the returned pgx.Row.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Exec executes a SQL command (INSERT, UPDATE, DELETE, DDL) and returns the command tag.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Tx defines transactional operations on the database.
//
// It provides methods to execute a function within a transaction context.
// The transaction is committed if the function completes successfully,
// otherwise it is rolled back. Panics inside the function also trigger rollback.
type Tx interface {
	// DoTx executes the given function inside a database transaction.
	// If fn returns nil, the transaction is committed.
	// If fn returns an error or panics, the transaction is rolled back.
	DoTx(ctx context.Context, opt pgx.TxOptions, fn func(tx RDBMS) error) error

	// DoTxContext is like DoTx, but also passes ctx to the function fn.
	// This allows passing deadlines, cancellation, or trace values into the transaction block.
	DoTxContext(ctx context.Context, opt pgx.TxOptions, fn func(ctx context.Context, tx RDBMS) error) error
}
