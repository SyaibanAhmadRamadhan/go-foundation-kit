package sqlx

import (
	"context"
	"database/sql"
	"time"
)

type rdbms struct {
	db    *sql.DB
	sc    *stmtCache
	tx    *sql.Tx
	hooks []DBHook
}

type rdbmsConfig struct {
	shardCount int
	hashFn     func(string, byte) string
	minCount   int64
	janIntv    time.Duration
	idleTTL    time.Duration
	ctx        context.Context
	hooks      []DBHook
}

// NewRDBMS constructs an RDBMS instance on top of *sql.DB with an internal
// prepared-statement cache and optional hooks for instrumentation.
// The stmt cache runs a background janitor if janIntv > 0 (using cfg.ctx).
func NewRDBMS(db *sql.DB, opts ...Option) *rdbms {
	cfg := defaultConfig()
	for _, o := range opts {
		o.apply(cfg)
	}

	sc := &stmtCache{
		shards:   make([]shard, cfg.shardCount),
		hashFn:   cfg.hashFn,
		minCount: cfg.minCount,
		janIntv:  cfg.janIntv,
		idleTTL:  cfg.idleTTL,
	}
	for i := range sc.shards {
		sc.shards[i] = shard{
			core:  make(map[string]*entry),
			queue: make(map[string]*entry),
			locks: make(map[string]*perKeyLock),
		}
	}

	if sc.janIntv > 0 {
		go sc.runJanitor(cfg.ctx)
	}

	return &rdbms{
		db:    db,
		sc:    sc,
		hooks: cfg.hooks,
	}
}

// QueryContext executes a query that returns rows, typically a SELECT.
// If r is inside a transaction, it uses that tx; otherwise it uses the base *sql.DB.
// Hook timing and error are recorded in HookInfo (OpQuery).
func (r *rdbms) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	info := &HookInfo{
		Op:       OpQuery,
		SQL:      query,
		Args:     args,
		InTx:     r.tx != nil,
		Prepared: false,
		Start:    time.Now(),
	}
	ctx = r.callBefore(ctx, info)

	var (
		rows *sql.Rows
		err  error
	)
	if r.tx != nil {
		rows, err = r.tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = r.db.QueryContext(ctx, query, args...)
	}
	info.Err = err
	info.End = time.Now()
	r.callAfter(ctx, info)
	return rows, err
}

// QueryRowContext executes a query that is expected to return at most one row.
// Errors from the underlying driver are usually reported at Scan time on the returned *sql.Row.
// Hook timing is recorded (OpQueryRow). Any immediate driver error will be set on HookInfo.Err.
func (r *rdbms) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	info := &HookInfo{
		Op:       OpQueryRow,
		SQL:      query,
		Args:     args,
		InTx:     r.tx != nil,
		Prepared: false,
		Start:    time.Now(),
	}
	ctx = r.callBefore(ctx, info)
	defer func() { info.End = time.Now(); r.callAfter(ctx, info) }()

	if r.tx != nil {
		return r.tx.QueryRowContext(ctx, query, args...)
	}
	return r.db.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a statement that does not return rows (INSERT/UPDATE/DELETE/DDL).
// If inside a transaction, the tx is used; otherwise the base *sql.DB is used.
// Hook timing and RowsAffected (if available) are recorded (OpExec).
func (r *rdbms) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	info := &HookInfo{
		Op:       OpExec,
		SQL:      query,
		Args:     args,
		InTx:     r.tx != nil,
		Prepared: false,
		Start:    time.Now(),
	}
	ctx = r.callBefore(ctx, info)
	defer func() { info.End = time.Now(); r.callAfter(ctx, info) }()

	var (
		res sql.Result
		err error
	)
	if r.tx != nil {
		res, err = r.tx.ExecContext(ctx, query, args...)
	} else {
		res, err = r.db.ExecContext(ctx, query, args...)
	}
	if err != nil {
		info.Err = err
		return nil, err
	}

	// Not all drivers support RowsAffected (e.g., some DDL). Ignore error if not supported.
	if ra, e2 := res.RowsAffected(); e2 == nil {
		v := ra
		info.Rows = &v
	}
	return res, nil
}

// PrepareContext creates a prepared statement on the base *sql.DB.
// NOTE: The tx-bound variant is commented out—if you want statements bound to the
// current transaction, you may switch to r.tx.PrepareContext when r.tx != nil.
// Hook timing and error are recorded (OpPrepare).
func (r *rdbms) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	info := &HookInfo{
		Op:    OpPrepare,
		SQL:   query,
		InTx:  r.tx != nil,
		Start: time.Now(),
	}
	ctx = r.callBefore(ctx, info)
	defer func() { info.End = time.Now(); r.callAfter(ctx, info) }()

	// If you need tx-scoped prepared statements, uncomment:
	// if r.tx != nil {
	// 	st, err := r.tx.PrepareContext(ctx, query)
	// 	info.Err = err
	// 	return st, err
	// }
	st, err := r.db.PrepareContext(ctx, query)
	info.Err = err
	return st, err
}

// DoTxContext runs fn within a database transaction.
// If r is already inside a transaction, fn is executed using the current receiver (no new tx).
// Otherwise, a child rdbms with tx-bound context is created. Commit/rollback events are
// surfaced via hooks (OpTxBegin, OpTxCommit, OpTxRollback).
//
// Behavior:
//   - Panic inside fn: transaction is rolled back, then panic is rethrown.
//   - fn returns error: transaction is rolled back, error is returned.
//   - fn returns nil: transaction is committed; commit error (if any) is returned.
func (r *rdbms) DoTxContext(ctx context.Context, opt *sql.TxOptions, fn func(ctx context.Context, tx RDBMS) error) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Already inside a transaction: run on the same receiver (nested logical tx, single physical tx).
	if r.tx != nil {
		return fn(ctx, r)
	}

	// Begin
	beg := &HookInfo{
		Op:    OpTxBegin,
		Start: time.Now(),
	}
	ctx = r.callBefore(ctx, beg)
	tx, err := r.db.BeginTx(ctx, opt)
	beg.Err, beg.End = err, time.Now()
	r.callAfter(ctx, beg)
	if err != nil {
		return err
	}

	// Child context bound to this tx
	child := &rdbms{db: r.db, sc: r.sc, tx: tx, hooks: r.hooks}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			roll := &HookInfo{Op: OpTxRollback, Err: err, Start: time.Now(), End: time.Now()}
			r.callAfter(ctx, roll)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			roll := &HookInfo{Op: OpTxRollback, Err: err, Start: time.Now(), End: time.Now()}
			r.callAfter(ctx, roll)
			return
		}
		cm := &HookInfo{Op: OpTxCommit, Start: time.Now()}
		cerr := tx.Commit()
		cm.Err, cm.End = cerr, time.Now()
		r.callAfter(ctx, cm)
	}()

	return fn(ctx, child)
}

// Close releases all cached statements across all shards and close db.
func (c *rdbms) Close() error {
	c.sc.close()
	return c.db.Close()
}

// callBefore executes all registered DBHook.Before in order, threading context through.
// Any hook may enrich the context (e.g., tracing IDs, timeouts).
func (r *rdbms) callBefore(ctx context.Context, info *HookInfo) context.Context {
	for _, h := range r.hooks {
		ctx = h.Before(ctx, info)
	}
	return ctx
}

// callAfter executes all registered DBHook.After in order.
// Hooks should be non-blocking and panic-safe at implementation site.
func (r *rdbms) callAfter(ctx context.Context, info *HookInfo) {
	for _, h := range r.hooks {
		h.After(ctx, info)
	}
}
