package sqlx

import (
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

// defaultConfig returns the default RDBMS configuration.
func defaultConfig() *rdbmsConfig {
	return &rdbmsConfig{
		hooks: make([]DBHook, 0),
	}
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

// UseDebug is a helper option to attach a DebugHook for logging SQL queries.
// If withArgs is true, query arguments will also be logged.
func UseDebug(withArgs bool) Option {
	return UseHook(&DebugHook{
		WithArgs: withArgs,
	})
}

type ObservabilityHookOption func(*ObservabilityHook)

// UseObservability is a helper option to attach an ObservabilityHook for SQL logs.
// By default it logs all SQL operations, without args, using a 500ms slow threshold.
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

// WithObservabilitySlowThreshold configures the minimum duration treated as a slow SQL operation.
func WithObservabilitySlowThreshold(d time.Duration) ObservabilityHookOption {
	return func(h *ObservabilityHook) {
		h.SlowThreshold = d
	}
}
