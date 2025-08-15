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

	if ra, e2 := res.RowsAffected(); e2 == nil {
		v := ra
		info.Rows = &v
	}
	return res, nil
}

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

	if ra, e2 := res.RowsAffected(); e2 == nil {
		v := ra
		info.Rows = &v
	}
	return res, nil
}

func (r *rdbms) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	info := &HookInfo{
		Op:    OpPrepare,
		SQL:   query,
		InTx:  r.tx != nil,
		Start: time.Now(),
	}
	ctx = r.callBefore(ctx, info)
	defer func() { info.End = time.Now(); r.callAfter(ctx, info) }()

	// if r.tx != nil {
	// 	st, err := r.tx.PrepareContext(ctx, query)
	// 	info.Err = err
	// 	return st, err
	// }
	st, err := r.db.PrepareContext(ctx, query)
	info.Err = err
	return st, err
}

func (r *rdbms) DoTxContext(ctx context.Context, opt *sql.TxOptions, fn func(ctx context.Context, tx RDBMS) error) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if r.tx != nil {
		return fn(ctx, r)
	}

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

func (r *rdbms) callBefore(ctx context.Context, info *HookInfo) context.Context {
	for _, h := range r.hooks {
		ctx = h.Before(ctx, info)
	}
	return ctx
}

func (r *rdbms) callAfter(ctx context.Context, info *HookInfo) {
	for _, h := range r.hooks {
		h.After(ctx, info)
	}
}
