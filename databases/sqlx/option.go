package sqlx

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"
)

// Option defines a configuration option for RDBMS.
// It uses the functional options pattern to customize rdbmsConfig.
type Option interface {
	apply(*rdbmsConfig)
}

// optFunc is an adapter to allow the use of ordinary functions as Option.
type optFunc func(*rdbmsConfig)

func (o optFunc) apply(r *rdbmsConfig) {
	o(r)
}

// hashKey generates a hashed string from the SQL query and a prefix byte
// using FNV-1a hashing. This is useful to create consistent, short keys
// for statement caching across shards.
func hashKey(sql string, prefix byte) string {
	h := fnv.New64a()
	h.Write([]byte{prefix})
	h.Write([]byte(sql))
	return fmt.Sprintf("%x", h.Sum64())
}

// defaultConfig returns the default RDBMS configuration for statement caching.
// Includes default shard count, hashing function, promotion threshold,
// janitor interval, idle TTL, and a background context.
func defaultConfig() *rdbmsConfig {
	return &rdbmsConfig{
		shardCount: 16,               // default number of shards in the statement cache
		hashFn:     hashKey,          // default key hashing function
		minCount:   2,                // minimum hits before promoting a stmt from queue → core
		janIntv:    30 * time.Second, // how often the janitor runs to clean idle statements
		idleTTL:    30 * time.Minute, // how long a statement can stay idle before eviction
		ctx:        context.Background(),
	}
}

// WithStmtShardCount sets the number of shards in the statement cache.
func WithStmtShardCount(n int) Option {
	return optFunc(func(c *rdbmsConfig) {
		if n <= 0 {
			n = 1
		}
		c.shardCount = n
	})
}

// WithStmtJanitorInterval sets how often the cache janitor runs.
func WithStmtJanitorInterval(d time.Duration) Option {
	return optFunc(func(c *rdbmsConfig) { c.janIntv = d })
}

// WithStmtIdleTTL sets the maximum idle time for cached statements
// before they are eligible for eviction.
func WithStmtIdleTTL(d time.Duration) Option {
	return optFunc(func(c *rdbmsConfig) { c.idleTTL = d })
}

// WithStmtMinCount sets the minimum number of accesses a statement must have
// before being promoted from the "queue" cache to the "core" cache.
func WithStmtMinCount(n int64) Option {
	return optFunc(func(c *rdbmsConfig) { c.minCount = n })
}

// WithStmtHashFn sets a custom hashing function for cache keys.
func WithStmtHashFn(fn func(string, byte) string) Option {
	return optFunc(func(c *rdbmsConfig) {
		if fn != nil {
			c.hashFn = fn
		}
	})
}

// WithStmtContext sets the context used by the statement cache janitor.
func WithStmtContext(ctx context.Context) Option {
	return optFunc(func(c *rdbmsConfig) {
		c.ctx = ctx
	})
}

// UseHook adds one or more DBHook instances to the RDBMS configuration.
// Hooks can be used for logging, tracing, metrics, etc.
func UseHook(h ...DBHook) Option {
	if len(h) > 0 {
		return optFunc(func(rc *rdbmsConfig) {
			if rc.hooks == nil {
				rc.hooks = make([]DBHook, 0)
			}
			rc.hooks = append(rc.hooks, h...)
		})
	}
	// No hooks provided → return a no-op option.
	return optFunc(func(rc *rdbmsConfig) {})
}

// UseDebugNql is a helper option to attach a DebugHook for logging SQL queries.
// If withArgs is true, query arguments will also be logged.
func UseDebugNql(withArgs bool) Option {
	return UseHook(&DebugHook{
		WithArgs: withArgs,
	})
}
