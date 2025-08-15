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

// rdbms is a PostgreSQL abstraction that implements RDBMS and Tx interfaces using pgx and squirrel.
type rdbms struct {
	db *pgxpool.Pool
	queryExecutor
	isTx bool
}

// NewRDBMS creates a new instance of rdbms using the given pgx connection pool.
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

func newRDBMSWithExecutor(db *pgxpool.Pool, executor queryExecutor) *rdbms {
	return &rdbms{
		db:            db,
		queryExecutor: executor,
		isTx:          true,
	}
}

// QuerySq executes a SELECT query built with squirrel and returns the result rows.
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

	err = fn(rows)

	return err
}

// ExecSq executes a write query (INSERT, UPDATE, DELETE) built with squirrel.
func (s *rdbms) ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return s.Exec(ctx, rawQuery, args...)
}

// QueryRowSq executes a SELECT query with squirrel and returns a single row.
func (s *rdbms) QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	return s.QueryRow(ctx, rawQuery, args...), nil
}

// QuerySqPagination executes a paginated SELECT query using squirrel and returns paginated result rows.
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

	err = s.QuerySq(ctx, query, fn)
	if err != nil {
		return primitive.PaginationOutput{}, err
	}

	return primitive.CreatePaginationOutput(paginationInput, totalData), nil
}

// DoTx executes a function within a database transaction.
// It commits the transaction if fn returns nil, otherwise rolls it back.
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

// DoTxContext is like DoTx but also passes the context to the transactional function.
func (s *rdbms) DoTxContext(
	ctx context.Context,
	opt pgx.TxOptions,
	fn func(ctx context.Context, tx RDBMS) (err error),
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
