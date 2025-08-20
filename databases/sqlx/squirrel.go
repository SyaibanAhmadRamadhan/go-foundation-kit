package sqlx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
)

// ExecSq executes a write query (INSERT/UPDATE/DELETE/DDL) built with Squirrel.
// If useStmt is true, the query is executed via prepared statement (ExecStmtContext),
// otherwise via direct ExecContext.
// Returns the sql.Result from the execution.
//
// Errors:
//   - wraps Squirrel builder errors as "failed parse squirrel".
//   - returns underlying Exec/ExecStmt errors.
func (r *rdbms) ExecSq(ctx context.Context, query squirrel.Sqlizer, useStmt bool) (sql.Result, error) {
	q, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed parse squirrel: %w", err)
	}

	if useStmt {
		return r.ExecStmtContext(ctx, q, args...)
	}
	return r.ExecContext(ctx, q, args...)
}

// QuerySq executes a SELECT query built with Squirrel and passes the resulting rows
// to the provided callback fn. The rows are always closed after fn returns.
// If useStmt is true, executes via prepared statement (QueryStmtContext),
// otherwise via direct QueryContext.
//
// Contract:
//   - fn is called exactly once iff the query succeeds and returns rows.
//   - rows.Close() is deferred here; fn MUST NOT close rows again.
//   - fn should fully iterate rows and handle scanning errors.
//
// Errors:
//   - wraps Squirrel builder errors as "failed parse squirrel".
//   - returns query errors or fn(rows) error.
func (r *rdbms) QuerySq(ctx context.Context, query squirrel.Sqlizer, useStmt bool, fn func(rows *sql.Rows) error) error {
	q, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed parse squirrel: %w", err)
	}

	if useStmt {
		rows, err := r.QueryStmtContext(ctx, q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		return fn(rows)
	}

	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return fn(rows)
}

// QuerySqPagination executes a paginated SELECT using two Squirrel builders:
//   - countQuery: SELECT COUNT(*) ... to compute total rows.
//   - query:      the actual SELECT with LIMIT/OFFSET applied.
//
// If useStmt is true, both queries are run via prepared statements.
// The provided fn receives the result rows from the main SELECT and must
// iterate/scan them. Rows are closed after fn returns.
//
// Returns:
//   - primitive.PaginationOutput computed from paginationInput and totalData.
//   - error from building queries, count/select execution, scanning COUNT(*), or fn.
//
// Contract:
//   - fn is called iff the SELECT query succeeds and returns rows.
//   - rows.Close() is handled here; fn MUST NOT close rows.
//
// Notes:
//   - countQuery should SELECT a single BIGINT/INT column as COUNT result.
func (r *rdbms) QuerySqPagination(
	ctx context.Context,
	countQuery, query squirrel.SelectBuilder,
	useStmt bool,
	paginationInput primitive.PaginationInput,
	fn func(rows *sql.Rows) error,
) (primitive.PaginationOutput, error) {
	query = query.Limit(uint64(paginationInput.PageSize)).Offset(uint64(primitive.GetOffsetValue(
		paginationInput.Page,
		paginationInput.PageSize,
	)))
	q, args, err := query.ToSql()
	if err != nil {
		return primitive.PaginationOutput{}, fmt.Errorf("failed parse squirrel: %w", err)
	}

	qCount, argsCount, err := countQuery.ToSql()
	if err != nil {
		return primitive.PaginationOutput{}, fmt.Errorf("failed parse squirrel: %w", err)
	}

	var totalData int64
	if useStmt {
		row, err := r.QueryRowStmtContext(ctx, qCount, argsCount...)
		if err != nil {
			return primitive.PaginationOutput{}, fmt.Errorf("failed count data: %w", err)
		}
		if err = row.Scan(&totalData); err != nil {
			return primitive.PaginationOutput{}, fmt.Errorf("failed count data: %w", err)
		}

		rows, err := r.QueryStmtContext(ctx, q, args...)
		if err != nil {
			return primitive.PaginationOutput{}, err
		}
		defer rows.Close()

		return primitive.CreatePaginationOutput(paginationInput, totalData), fn(rows)
	}

	row := r.QueryRowContext(ctx, qCount, argsCount...)
	if err = row.Scan(&totalData); err != nil {
		return primitive.PaginationOutput{}, fmt.Errorf("failed count data: %w", err)
	}

	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return primitive.PaginationOutput{}, err
	}
	defer rows.Close()

	return primitive.CreatePaginationOutput(paginationInput, totalData), fn(rows)
}

// QueryRowSq executes a SELECT (single-row) using a Squirrel builder.
// If useStmt is true, executes via prepared statement (QueryRowStmtContext),
// otherwise via direct QueryRowContext.
//
// Returns:
//   - *sql.Row that caller must Scan().
//   - error for build or prepared-statement acquisition; QueryRowContext itself
//     defers errors to Scan() (so error may be seen at Scan time).
func (r *rdbms) QueryRowSq(ctx context.Context, query squirrel.Sqlizer, useStmt bool) (*sql.Row, error) {
	q, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed parse squirrel: %w", err)
	}

	if useStmt {
		row, err := r.QueryRowStmtContext(ctx, q, args...)
		if err != nil {
			return nil, err
		}
		return row, nil
	}

	row := r.QueryRowContext(ctx, q, args...)
	return row, nil
}
