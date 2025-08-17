package sqlx

import (
	"context"
	"database/sql"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases"
)

// entry represents a cached prepared statement.
// It tracks usage, hit count, and last access time for eviction and promotion.
type entry struct {
	stmt       *sql.Stmt
	inUse      atomic.Int32 // Number of concurrent usages (prevents closing while in use)
	count      atomic.Int64 // Total number of times this statement has been used
	lastAccess atomic.Int64 // Last access timestamp in UnixNano
}

// perKeyLock ensures that only one goroutine prepares a given statement at a time.
type perKeyLock struct{ mu sync.Mutex }

// shard stores two maps for statement caching:
// - core: "hot" statements that are frequently used
// - queue: recently prepared statements waiting to be promoted to core
type shard struct {
	mu    sync.RWMutex
	core  map[string]*entry
	queue map[string]*entry
	locks map[string]*perKeyLock
}

// stmtCache manages the sharded prepared statement cache.
type stmtCache struct {
	shards   []shard
	hashFn   func(string, byte) string // Hash function to generate cache keys
	minCount int64                     // Minimum usage count to promote from queue → core
	janIntv  time.Duration             // How often the janitor runs
	idleTTL  time.Duration             // How long a statement can be idle before eviction
}

// withKeyLock executes a function under a per-statement key lock
// to ensure that the same statement is not prepared multiple times concurrently.
func (s *shard) withKeyLock(key string, fn func() (*entry, error)) (*entry, error) {
	s.mu.Lock()
	lk, ok := s.locks[key]
	if !ok {
		lk = &perKeyLock{}
		s.locks[key] = lk
	}
	s.mu.Unlock()

	lk.mu.Lock()
	defer lk.mu.Unlock()
	return fn()
}

// getShard returns the shard responsible for the given cache key.
func (c *stmtCache) getShard(key string) *shard {
	i := fnvIndex(key, len(c.shards))
	return &c.shards[i]
}

// getOrPrepare returns a cached prepared statement if available,
// otherwise it prepares a new one and stores it in the queue.
// This method is transaction-aware (via queryExecutor interface).
func (c *stmtCache) getOrPrepare(ctx context.Context, db queryExecutor, sql string) (*entry, error) {
	key := c.hashFn(databases.NormalizeSQL(sql), 'q')
	s := c.getShard(key)

	s.mu.RLock()
	if e := s.core[key]; e != nil {
		s.mu.RUnlock()
		e.inUse.Add(1)
		e.count.Add(1)
		return e, nil
	}
	if e := s.queue[key]; e != nil {
		s.mu.RUnlock()
		e.inUse.Add(1)
		e.count.Add(1)
		return e, nil
	}
	s.mu.RUnlock()

	return s.withKeyLock(key, func() (*entry, error) {
		s.mu.RLock()
		if e := s.core[key]; e != nil {
			s.mu.RUnlock()
			e.inUse.Add(1)
			e.count.Add(1)
			return e, nil
		}
		if e := s.queue[key]; e != nil {
			s.mu.RUnlock()
			e.inUse.Add(1)
			e.count.Add(1)
			return e, nil
		}
		s.mu.RUnlock()

		st, err := db.PrepareContext(ctx, sql)
		if err != nil {
			return nil, err
		}

		e := &entry{stmt: st}
		e.inUse.Store(1)
		e.count.Store(1)
		e.lastAccess.Store(time.Now().UnixNano())

		s.mu.Lock()
		s.queue[key] = e
		s.mu.Unlock()
		return e, nil
	})
}

// put releases an entry after use and updates its last access time.
func (c *stmtCache) put(e *entry) {
	e.inUse.Add(-1)
	e.lastAccess.Store(time.Now().UnixNano())
}

// promoteIfHot moves a statement from queue → core if it has enough usage count.
func (c *stmtCache) promoteIfHot(s *shard, key string) {
	s.mu.Lock()
	if e := s.queue[key]; e != nil && e.count.Load() > c.minCount {
		delete(s.queue, key)
		s.core[key] = e
	}
	s.mu.Unlock()
}

// evictIdleCore removes idle statements from the core cache that
// have not been used for longer than idleTTL.
func (c *stmtCache) evictIdleCore(s *shard, now int64) int {
	if c.idleTTL <= 0 {
		return 0
	}
	evicted := 0
	ttl := int64(c.idleTTL)

	for k, e := range s.core {
		if e.inUse.Load() == 0 && (now-e.lastAccess.Load()) > ttl {
			if e.stmt != nil {
				_ = e.stmt.Close()
			}
			delete(s.core, k)
			evicted++
		}
	}
	return evicted
}

// evictIdleQueue removes idle statements from the queue cache that
// have not been used for longer than idleTTL.
func (c *stmtCache) evictIdleQueue(s *shard, now int64) int {
	if c.idleTTL <= 0 {
		return 0
	}
	evicted := 0
	ttl := int64(c.idleTTL)

	for k, e := range s.queue {
		if e.inUse.Load() == 0 && (now-e.lastAccess.Load()) > ttl {
			if e.stmt != nil {
				_ = e.stmt.Close()
			}
			delete(s.core, k)
			evicted++
		}
	}
	return evicted
}

// runJanitor periodically promotes hot statements and evicts idle ones.
// Runs in a background goroutine until the context is cancelled.
func (c *stmtCache) runJanitor(ctx context.Context) {
	t := time.NewTicker(c.janIntv)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			// c.close()
			return
		case <-t.C:
			for i := range c.shards {
				s := &c.shards[i]

				s.mu.RLock()
				keys := make([]string, 0, len(s.queue))
				for k := range s.queue {
					keys = append(keys, k)
				}
				s.mu.RUnlock()
				for _, k := range keys {
					c.promoteIfHot(s, k)
				}

				s.mu.Lock()
				for k, e := range s.queue {
					if e.inUse.Load() == 0 {
						if e.stmt != nil {
							_ = e.stmt.Close()
						}
						delete(s.queue, k)
					} else {
						e.count.Store(e.count.Load() >> 1)
					}
				}
				_ = c.evictIdleCore(s, time.Now().Unix())
				_ = c.evictIdleQueue(s, time.Now().Unix())
				s.mu.Unlock()
			}
		}
	}
}

// Close releases all cached statements across all shards.
func (c *stmtCache) close() {
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		for k, e := range s.core {
			if e.stmt != nil {
				_ = e.stmt.Close()
			}
			delete(s.core, k)
		}
		for k, e := range s.queue {
			if e.stmt != nil {
				_ = e.stmt.Close()
			}
			delete(s.queue, k)
		}
		s.mu.Unlock()
	}
}

// fnvIndex returns a shard index for a given key using FNV-1a hash.
func fnvIndex(key string, shardCount int) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % uint32(shardCount))
}
