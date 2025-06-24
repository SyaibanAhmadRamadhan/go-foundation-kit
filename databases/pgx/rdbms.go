package libpgx

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RDBMS interface {
	ReadQuery
	WriterCommand
	queryExecutor
}

type WriterCommand interface {
	WriterCommandSquirrel
}

type ReadQuery interface {
	ReadQuerySquirrel
}

type WriterCommandSquirrel interface {
	ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error)
}

type ReadQuerySquirrel interface {
	QuerySq(ctx context.Context, query squirrel.Sqlizer) (pgx.Rows, error)
	QuerySqPagination(ctx context.Context, countQuery, query squirrel.SelectBuilder, paginationInput primitive.PaginationInput) (
		pgx.Rows, primitive.PaginationOutput, error)
	QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error)
}

type queryExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Tx interface {
	DoTx(ctx context.Context, opt *sql.TxOptions, fn func(tx RDBMS) error) error
	DoTxContext(ctx context.Context, opt *sql.TxOptions, fn func(ctx context.Context, tx RDBMS) error) error
}
