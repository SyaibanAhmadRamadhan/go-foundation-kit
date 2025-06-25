package libpgx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/pgx/otelpgx"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPgxPoolWithOtel(connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithIncludeQueryParameters(),
		otelpgx.WithTrimSQLInSpanName(),
	)
	conn, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return conn, err
}

type rdbms struct {
	db *pgxpool.Pool
	queryExecutor
	isTx bool
}

var _ RDBMS = (*rdbms)(nil)
var _ Tx = (*rdbms)(nil)

func NewRDBMS(db *pgxpool.Pool) *rdbms {
	return &rdbms{
		db:            db,
		queryExecutor: db,
	}
}

func (s *rdbms) QuerySq(ctx context.Context, query squirrel.Sqlizer) (pgx.Rows, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	res, err := s.Query(ctx, rawQuery, args...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *rdbms) ExecSq(ctx context.Context, query squirrel.Sqlizer) (pgconn.CommandTag, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	res, err := s.Exec(ctx, rawQuery, args...)
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	return res, nil
}

func (s *rdbms) QueryRowSq(ctx context.Context, query squirrel.Sqlizer) (pgx.Row, error) {
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := s.QueryRow(ctx, rawQuery, args...)

	return row, nil
}

func (s *rdbms) QuerySqPagination(ctx context.Context, countQuery, query squirrel.SelectBuilder, paginationInput primitive.PaginationInput) (
	pgx.Rows, primitive.PaginationOutput, error) {

	pageSize := paginationInput.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := max(primitive.GetOffsetValue(paginationInput.Page, pageSize), 0)

	query = query.Limit(uint64(pageSize))
	query = query.Offset(uint64(offset))

	totalData := int64(0)
	row, err := s.QueryRowSq(ctx, countQuery)
	if err != nil {
		return nil, primitive.PaginationOutput{}, err
	}

	err = row.Scan(&totalData)
	if err != nil {
		return nil, primitive.PaginationOutput{}, err
	}

	rows, err := s.QuerySq(ctx, query)
	if err != nil {
		return nil, primitive.PaginationOutput{}, err
	}

	return rows, primitive.CreatePaginationOutput(paginationInput, totalData), nil
}

func (s *rdbms) injectTx(tx pgx.Tx) *rdbms {
	newRdbms := *s
	newRdbms.queryExecutor = tx
	newRdbms.isTx = true
	return &newRdbms
}

func (s *rdbms) DoTx(ctx context.Context, opt pgx.TxOptions, fn func(tx RDBMS) (err error)) (err error) {
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
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				if !errors.Is(err, sql.ErrTxDone) {
					err = errors.Join(err, errRollback)
				}
			}
		} else {
			if errCommit := tx.Commit(ctx); errCommit != nil {
				if !errors.Is(err, sql.ErrTxDone) {
					err = errors.Join(err, errCommit)
				}
			}
		}
	}()

	err = fn(s.injectTx(tx))
	return
}

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
		switch {
		case recover() != nil:
			_ = tx.Rollback(ctx)
			panic("panic occurred")
		case err != nil:
			if errRollback := tx.Rollback(ctx); errRollback != nil && !errors.Is(err, sql.ErrTxDone) {
				err = errors.Join(err, errRollback)
			}
		default:
			if errCommit := tx.Commit(ctx); errCommit != nil && !errors.Is(errCommit, sql.ErrTxDone) {
				err = errors.Join(err, errCommit)
			}
		}
	}()

	err = fn(ctx, s.injectTx(tx))
	return
}
