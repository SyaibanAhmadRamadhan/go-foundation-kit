package pgxx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	db    *pgxpool.Pool
	hooks []DBHook
	queryExecutor
	isTx bool
}

// NewRDBMS creates a new RDBMS backed by a pgx connection pool.
func NewRDBMS(conn string, opts ...Option) (*rdbms, func(), error) {
	poolCfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection pool: %w", err)
	}

	internalCfg := defaultConfig(poolCfg)
	for _, o := range opts {
		o.apply(internalCfg)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), internalCfg.pool)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	return &rdbms{
		db:            db,
		queryExecutor: db,
		hooks:         internalCfg.hooks,
	}, db.Close, nil
}

// newRDBMSWithExecutor creates a transactional RDBMS that uses the given executor.
// It is intended for use inside transactions (e.g., pgx.Tx).
func newRDBMSWithExecutor(db *pgxpool.Pool, executor queryExecutor, hooks []DBHook) *rdbms {
	return &rdbms{
		db:            db,
		queryExecutor: executor,
		hooks:         hooks,
		isTx:          true,
	}
}

func (s *rdbms) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	info := &HookInfo{Op: OpQuery, SQL: sql, Args: args, InTx: s.isTx, Start: time.Now()}
	ctx = s.callBefore(ctx, info)

	rows, err := s.queryExecutor.Query(ctx, sql, args...)
	info.Err = err
	info.End = time.Now()
	s.callAfter(ctx, info)
	return rows, err
}

func (s *rdbms) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	info := &HookInfo{Op: OpQueryRow, SQL: sql, Args: args, InTx: s.isTx, Start: time.Now()}
	ctx = s.callBefore(ctx, info)
	defer func() {
		info.End = time.Now()
		s.callAfter(ctx, info)
	}()

	return s.queryExecutor.QueryRow(ctx, sql, args...)
}

func (s *rdbms) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	info := &HookInfo{Op: OpExec, SQL: sql, Args: arguments, InTx: s.isTx, Start: time.Now()}
	ctx = s.callBefore(ctx, info)

	tag, err := s.queryExecutor.Exec(ctx, sql, arguments...)
	info.Err = err
	if err == nil {
		rowsAffected := tag.RowsAffected()
		info.Rows = &rowsAffected
	}
	info.End = time.Now()
	s.callAfter(ctx, info)
	return tag, err
}

// QuerySq executes a SELECT query built with Squirrel and invokes the provided
// callback with the result rows.
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
func (s *rdbms) ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return s.Exec(ctx, rawQuery, args...)
}

// QueryRowSq executes a SELECT query built with Squirrel and returns a single row.
func (s *rdbms) QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	return s.QueryRow(ctx, rawQuery, args...), nil
}

func (s *rdbms) GetDB() *pgxpool.Pool {
	return s.db
}

// QuerySqPagination executes a paginated SELECT query built with Squirrel.
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
func (s *rdbms) DoTx(ctx context.Context, opt pgx.TxOptions, fn func(tx RDBMS) error) (err error) {
	if opt.IsoLevel == "" {
		opt = pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite}
	}

	beg := &HookInfo{Op: OpTxBegin, InTx: true}
	ctx = s.callBefore(ctx, beg)
	tx, err := s.db.BeginTx(ctx, opt)
	beg.Err = err
	beg.End = time.Now()
	s.callAfter(ctx, beg)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			roll := &HookInfo{Op: OpTxRollback, InTx: true, Err: fmt.Errorf("panic: %v", p)}
			ctx = s.callBefore(ctx, roll)
			_ = tx.Rollback(ctx)
			roll.End = time.Now()
			s.callAfter(ctx, roll)
			panic(p)
		}

		if err != nil {
			roll := &HookInfo{Op: OpTxRollback, InTx: true, Err: err}
			ctx = s.callBefore(ctx, roll)
			if errRollback := tx.Rollback(ctx); errRollback != nil && !errors.Is(err, sql.ErrTxDone) {
				err = errors.Join(err, errRollback)
			}
			roll.Err = err
			roll.End = time.Now()
			s.callAfter(ctx, roll)
			return
		}

		cm := &HookInfo{Op: OpTxCommit, InTx: true}
		ctx = s.callBefore(ctx, cm)
		if errCommit := tx.Commit(ctx); errCommit != nil && !errors.Is(errCommit, sql.ErrTxDone) {
			err = errors.Join(err, errCommit)
		}
		cm.Err = err
		cm.End = time.Now()
		s.callAfter(ctx, cm)
	}()

	return fn(newRDBMSWithExecutor(s.db, tx, s.hooks))
}

// DoTxContext is like DoTx, but passes ctx along to the transactional function.
func (s *rdbms) DoTxContext(
	ctx context.Context,
	opt pgx.TxOptions,
	fn func(ctx context.Context, tx RDBMS) error,
) (err error) {
	if opt.IsoLevel == "" {
		opt = pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite}
	}

	beg := &HookInfo{Op: OpTxBegin, InTx: true}
	ctx = s.callBefore(ctx, beg)
	tx, err := s.db.BeginTx(ctx, opt)
	beg.Err = err
	beg.End = time.Now()
	s.callAfter(ctx, beg)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			roll := &HookInfo{Op: OpTxRollback, InTx: true, Err: fmt.Errorf("panic: %v", p)}
			ctx = s.callBefore(ctx, roll)
			_ = tx.Rollback(ctx)
			roll.End = time.Now()
			s.callAfter(ctx, roll)
			panic(p)
		}

		if err != nil {
			roll := &HookInfo{Op: OpTxRollback, InTx: true, Err: err}
			ctx = s.callBefore(ctx, roll)
			if errRollback := tx.Rollback(ctx); errRollback != nil && !errors.Is(err, sql.ErrTxDone) {
				err = errors.Join(err, errRollback)
			}
			roll.Err = err
			roll.End = time.Now()
			s.callAfter(ctx, roll)
			return
		}

		cm := &HookInfo{Op: OpTxCommit, InTx: true}
		ctx = s.callBefore(ctx, cm)
		if errCommit := tx.Commit(ctx); errCommit != nil && !errors.Is(errCommit, sql.ErrTxDone) {
			err = errors.Join(err, errCommit)
		}
		cm.Err = err
		cm.End = time.Now()
		s.callAfter(ctx, cm)
	}()

	return fn(ctx, newRDBMSWithExecutor(s.db, tx, s.hooks))
}

func (s *rdbms) callBefore(ctx context.Context, info *HookInfo) context.Context {
	for _, h := range s.hooks {
		ctx = h.Before(ctx, info)
	}
	return ctx
}

func (s *rdbms) callAfter(ctx context.Context, info *HookInfo) {
	for _, h := range s.hooks {
		h.After(ctx, info)
	}
}
