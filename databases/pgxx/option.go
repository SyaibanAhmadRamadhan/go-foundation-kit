package pgxx

import (
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option interface {
	apply(*rdbmsConfig)
}

type optFunc func(*rdbmsConfig)

type rdbmsConfig struct {
	pool  *pgxpool.Config
	hooks []DBHook
}

func defaultConfig(pool *pgxpool.Config) *rdbmsConfig {
	return &rdbmsConfig{
		pool:  pool,
		hooks: make([]DBHook, 0),
	}
}

func (o optFunc) apply(cfg *rdbmsConfig) {
	o(cfg)
}

func WithOtel(opts ...otelpgx.Option) Option {
	return optFunc(func(cfg *rdbmsConfig) {
		opts = append(opts, otelpgx.WithTrimSQLInSpanName())
		cfg.pool.ConnConfig.Tracer = otelpgx.NewTracer(
			opts...,
		)
	})
}

// UseHook adds one or more hooks to the pgxx RDBMS.
func UseHook(h ...DBHook) Option {
	if len(h) == 0 {
		return optFunc(func(cfg *rdbmsConfig) {})
	}

	return optFunc(func(cfg *rdbmsConfig) {
		cfg.hooks = append(cfg.hooks, h...)
	})
}

// UseDebug enables a simple SQL log hook.
func UseDebug(withArgs bool) Option {
	return UseHook(&DebugHook{WithArgs: withArgs})
}

type ObservabilityHookOption func(*ObservabilityHook)

// UseObservability enables zerolog-based SQL observability logs.
func UseObservability(opts ...ObservabilityHookOption) Option {
	hook := &ObservabilityHook{
		Mode:          ObservabilityLogAll,
		SlowThreshold: defaultObservabilitySlowThreshold,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(hook)
		}
	}

	return UseHook(hook)
}

// WithObservabilityArgs configures whether SQL args are included in observability logs.
func WithObservabilityArgs(withArgs bool) ObservabilityHookOption {
	return func(h *ObservabilityHook) {
		h.WithArgs = withArgs
	}
}

// WithObservabilityMode configures which SQL operations are emitted to observability logs.
func WithObservabilityMode(mode ObservabilityLogMode) ObservabilityHookOption {
	return func(h *ObservabilityHook) {
		h.Mode = mode
	}
}

// WithObservabilitySlowThreshold configures the minimum duration treated as slow.
func WithObservabilitySlowThreshold(d time.Duration) ObservabilityHookOption {
	return func(h *ObservabilityHook) {
		h.SlowThreshold = d
	}
}
