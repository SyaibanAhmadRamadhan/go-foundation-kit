package pgxx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// rdbms is a PostgreSQL abstraction that implements RDBMS and Tx
// interfaces using pgx and Squirrel. It delegates query execution
// to a pgx connection pool or transaction.
type rdbms struct {
	db *pgxpool.Pool
	queryExecutor
	isTx bool
}

// NewRDBMS creates a new RDBMS backed by a pgx connection pool.
//
// conn should be a PostgreSQL DSN string. Options (opts) can be
// provided to customize the pgx pool configuration.
//
// Returns the RDBMS instance, a cleanup function to close the pool,
// and an error if initialization fails.
func NewRDBMS(conn string, opts ...Option) (*rdbms, func(), error) {
	cfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection pool: %w", err)
	}
	for _, o := range opts {
		o.apply(cfg)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	return &rdbms{
		db:            db,
		queryExecutor: db,
	}, db.Close, nil
}

// newRDBMSWithExecutor creates a transactional RDBMS that uses the given executor.
// It is intended for use inside transactions (e.g., pgx.Tx).
func newRDBMSWithExecutor(db *pgxpool.Pool, executor queryExecutor) *rdbms {
	return &rdbms{
		db:            db,
		queryExecutor: executor,
		isTx:          true,
	}
}

// QuerySq executes a SELECT query built with Squirrel and invokes the provided
// callback with the result rows.
//
// The caller must consume rows within fn. Rows are closed automatically
// after fn returns.
func (s *rdbms) QuerySq(ctx context.Context, query squirrel.Sqlizer, fn func(rows pgx.Rows) error) error {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return err
	}

	rows, err := s.Query(ctx, rawQuery, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return fn(rows)
}

// ExecSq executes a write query (INSERT, UPDATE, DELETE) built with Squirrel.
//
// Returns the pgconn.CommandTag, which contains the number of rows affected
// and command information.
func (s *rdbms) ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return s.Exec(ctx, rawQuery, args...)
}

// QueryRowSq executes a SELECT query built with Squirrel and returns a single row.
//
// Errors from the driver are typically returned when calling Scan on the row.
func (s *rdbms) QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	return s.QueryRow(ctx, rawQuery, args...), nil
}

// QuerySqPagination executes a paginated SELECT query built with Squirrel.
//
// It uses countQuery to retrieve the total number of records,
// and query to fetch the paginated subset. paginationInput defines
// the page and page size. The fn callback is invoked with the paginated rows.
//
// Returns a PaginationOutput describing pagination metadata
// and any error encountered.
func (s *rdbms) QuerySqPagination(
	ctx context.Context,
	countQuery, query squirrel.SelectBuilder,
	paginationInput primitive.PaginationInput,
	fn func(rows pgx.Rows) error,
) (primitive.PaginationOutput, error) {
	pageSize := paginationInput.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := max(primitive.GetOffsetValue(paginationInput.Page, pageSize), 0)

	query = query.Limit(uint64(pageSize)).Offset(uint64(offset))

	// Get total count
	totalData := int64(0)
	row, err := s.QueryRowSq(ctx, countQuery)
	if err != nil {
		return primitive.PaginationOutput{}, err
	}
	if err := row.Scan(&totalData); err != nil {
		return primitive.PaginationOutput{}, err
	}

	if err = s.QuerySq(ctx, query, fn); err != nil {
		return primitive.PaginationOutput{}, err
	}

	return primitive.CreatePaginationOutput(paginationInput, totalData), nil
}

// DoTx executes a function within a database transaction.
//
// If fn returns nil, the transaction is committed. If fn returns
// an error or a panic occurs, the transaction is rolled back.
// Any rollback/commit error is joined with the returned error.
func (s *rdbms) DoTx(ctx context.Context, opt pgx.TxOptions, fn func(tx RDBMS) error) (err error) {
	if opt.IsoLevel == "" {
		opt = pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}
	}

	tx, err := s.db.BeginTx(ctx, opt)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil && !errors.Is(err, sql.ErrTxDone) {
				err = errors.Join(err, errRollback)
			}
		} else {
			if errCommit := tx.Commit(ctx); errCommit != nil && !errors.Is(errCommit, sql.ErrTxDone) {
				err = errors.Join(err, errCommit)
			}
		}
	}()

	return fn(newRDBMSWithExecutor(s.db, tx))
}

// DoTxContext is like DoTx, but passes ctx along to the transactional function.
//
// The transaction is committed if fn returns nil. If fn returns an error
// or a panic occurs, the transaction is rolled back. Any rollback/commit
// error is joined with the returned error.
func (s *rdbms) DoTxContext(
	ctx context.Context,
	opt pgx.TxOptions,
	fn func(ctx context.Context, tx RDBMS) error,
) (err error) {
	if opt.IsoLevel == "" {
		opt = pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}
	}

	tx, err := s.db.BeginTx(ctx, opt)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic("panic occurred")
		} else if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil && !errors.Is(err, sql.ErrTxDone) {
				err = errors.Join(err, errRollback)
			}
		} else {
			if errCommit := tx.Commit(ctx); errCommit != nil && !errors.Is(errCommit, sql.ErrTxDone) {
				err = errors.Join(err, errCommit)
			}
		}
	}()

	return fn(ctx, newRDBMSWithExecutor(s.db, tx))
}
