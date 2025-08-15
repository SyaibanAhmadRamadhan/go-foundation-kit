package sqlx

import (
	"context"
	"time"
)

// Op represents the type of database operation being executed.
// This can be used by hooks to differentiate logging or metrics by operation type.
type Op string

// Possible Op values for different SQL operations.
const (
	OpQuery      Op = "query"       // Query returning multiple rows
	OpQueryRow   Op = "query_row"   // Query returning a single row
	OpExec       Op = "exec"        // Executing a statement without returning rows
	OpPrepare    Op = "prepare"     // Preparing a SQL statement
	OpTxBegin    Op = "tx_begin"    // Beginning a transaction
	OpTxCommit   Op = "tx_commit"   // Committing a transaction
	OpTxRollback Op = "tx_rollback" // Rolling back a transaction
)

// HookInfo contains detailed information about a database operation,
// passed to hooks before and after the operation is executed.
type HookInfo struct {
	Op       Op        // The type of operation (query, exec, prepare, transaction)
	SQL      string    // The SQL query string
	Args     []any     // Query arguments, if any
	InTx     bool      // True if the operation is executed inside a transaction
	Prepared bool      // True if executed using a prepared statement or cache
	CacheHit *bool     // Optional: true if the prepared statement was retrieved from cache, false if newly prepared
	Start    time.Time // Start time of the operation (set in Before hook)
	End      time.Time // End time of the operation (set in After hook)
	Err      error     // Any error returned from the operation
	Rows     *int64    // Optional: number of rows affected (Exec) or returned (Query)
}

// DBHook defines the interface for database hooks.
// Hooks can be used for logging, tracing, or metrics collection.
//
// The sequence for hook calls:
//  1. Before(ctx, info) is called before the database operation starts.
//     - Can modify context or enrich HookInfo.
//  2. After(ctx, info) is called after the database operation ends.
//     - Has access to execution results, duration, and any errors.
type DBHook interface {
	// Before is called before the SQL operation begins.
	// Can modify and return a new context.
	Before(ctx context.Context, info *HookInfo) context.Context

	// After is called after the SQL operation ends.
	// Has access to the same HookInfo, now with End time and possible error.
	After(ctx context.Context, info *HookInfo)
}
