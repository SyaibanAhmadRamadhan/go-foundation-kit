package sqlx

import (
	"context"
	"database/sql"
	"time"
)

// QueryStmtContext executes a prepared statement that returns multiple rows.
// The statement is fetched from the internal stmt cache (prepare-if-missing)
// and returned to the cache after use.
//
// Behavior:
//   - Uses tx-bound statement when inside a transaction (r.tx != nil) via r.tx.StmtContext.
//   - Caller is responsible to Close() the returned *sql.Rows.
//   - Hooks are invoked with OpQuery and Prepared=true.
//
// Errors:
//   - Returns error from stmt cache (prepare/get) or QueryContext.
func (r *rdbms) QueryStmtContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	info := &HookInfo{
		Op:       OpQuery,
		SQL:      query,
		Args:     args,
		InTx:     r.tx != nil,
		Prepared: true,
		Start:    time.Now(),
	}
	ctx = r.callBefore(ctx, info)

	e, err := r.sc.getOrPrepare(ctx, r, query)
	if err != nil {
		info.Err = err
		info.End = time.Now()
		r.callAfter(ctx, info)
		return nil, err
	}
	defer r.sc.put(e)

	var rows *sql.Rows
	if r.tx != nil {
		rows, err = r.tx.StmtContext(ctx, e.stmt).QueryContext(ctx, args...)
	} else {
		rows, err = e.stmt.QueryContext(ctx, args...)
	}
	info.Err = err
	info.End = time.Now()
	r.callAfter(ctx, info)
	return rows, err
}

// QueryRowStmtContext executes a prepared statement expected to return a single row.
// The statement is fetched from the internal stmt cache and returned to the cache after use.
//
// Behavior:
//   - Uses tx-bound statement when inside a transaction via r.tx.StmtContext.
//   - Errors from drivers are typically surfaced on row.Scan(...).
//   - Hooks are invoked with OpQueryRow and Prepared=true.
//
// Returns:
//   - *sql.Row which the caller must Scan().
//   - An immediate error only if stmt retrieval/preparation fails.
func (r *rdbms) QueryRowStmtContext(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	info := &HookInfo{
		Op:       OpQueryRow,
		SQL:      query,
		Args:     args,
		InTx:     r.tx != nil,
		Prepared: true,
		Start:    time.Now(),
	}
	ctx = r.callBefore(ctx, info)
	defer func() { info.End = time.Now(); r.callAfter(ctx, info) }()

	e, err := r.sc.getOrPrepare(ctx, r, query)
	if err != nil {
		info.Err = err
		return nil, err
	}
	defer r.sc.put(e)

	if r.tx != nil {
		return r.tx.StmtContext(ctx, e.stmt).QueryRowContext(ctx, args...), nil
	}
	return e.stmt.QueryRowContext(ctx, args...), nil
}

// ExecStmtContext executes a prepared statement that does not return rows
// (INSERT/UPDATE/DELETE/DDL). The statement is fetched from the internal stmt
// cache and returned to the cache after use.
//
// Behavior:
//   - Uses tx-bound statement when inside a transaction via r.tx.StmtContext.
//   - Records RowsAffected() in HookInfo.Rows when supported by the driver.
//   - Hooks are invoked with OpExec and Prepared=true.
//
// Errors:
//   - Returns error from stmt cache (prepare/get) or ExecContext.
func (r *rdbms) ExecStmtContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	info := &HookInfo{
		Op:       OpExec,
		SQL:      query,
		Args:     args,
		InTx:     r.tx != nil,
		Prepared: true,
		Start:    time.Now(),
	}
	ctx = r.callBefore(ctx, info)
	defer func() { info.End = time.Now(); r.callAfter(ctx, info) }()

	e, err := r.sc.getOrPrepare(ctx, r, query)
	if err != nil {
		info.Err = err
		return nil, err
	}
	defer r.sc.put(e)

	var res sql.Result
	if r.tx != nil {
		res, err = r.tx.StmtContext(ctx, e.stmt).ExecContext(ctx, args...)
	} else {
		res, err = e.stmt.ExecContext(ctx, args...)
	}
	if err != nil {
		info.Err = err
		return nil, err
	}

	// Not all drivers support RowsAffected; ignore the error if not supported.
	if ra, e2 := res.RowsAffected(); e2 == nil {
		v := ra
		info.Rows = &v
	}
	return res, nil
}
