package libpgx

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// RDBMS is a high-level interface that combines read and write capabilities
// for interacting with a PostgreSQL database using pgx and Squirrel.
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

// WriterCommandSquirrel allows executing Squirrel SQL builders for write operations.
type WriterCommandSquirrel interface {
	// ExecSq executes a Squirrel-built write query.
	ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error)
}

// ReadQuerySquirrel provides read operations using Squirrel's SQL builder.
type ReadQuerySquirrel interface {
	// QuerySq executes a SELECT query using Squirrel and returns multiple rows.
	QuerySq(ctx context.Context, query squirrel.Sqlizer) (pgx.Rows, error)

	// QuerySqPagination executes paginated SELECT queries.
	QuerySqPagination(ctx context.Context, countQuery, query squirrel.SelectBuilder, paginationInput primitive.PaginationInput) (
		pgx.Rows, primitive.PaginationOutput, error)

	// QueryRowSq executes a SELECT query and returns a single row.
	QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error)
}

// queryExecutor provides low-level query execution using raw SQL strings.
// This is useful for situations where Squirrel is not used.
type queryExecutor interface {
	// Query executes a SQL string and returns multiple rows.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow executes a SQL string and returns a single row.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Exec executes a SQL command and returns the result tag.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// Tx defines transactional operations on the database.
// Provides methods to execute blocks of code within a transaction context.
type Tx interface {
	// DoTx runs the given function inside a database transaction.
	// The transaction is committed if the function returns nil, otherwise rolled back.
	DoTx(ctx context.Context, opt pgx.TxOptions, fn func(tx RDBMS) error) error

	// DoTxContext is a context-aware version of DoTx that also passes the context to the callback.
	DoTxContext(ctx context.Context, opt pgx.TxOptions, fn func(ctx context.Context, tx RDBMS) (err error)) (err error)
}
