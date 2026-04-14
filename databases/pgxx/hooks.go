package pgxx

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability"
	"github.com/rs/zerolog"
)

// Op represents the type of database operation being executed.
type Op string

const (
	OpQuery      Op = "query"
	OpQueryRow   Op = "query_row"
	OpExec       Op = "exec"
	OpTxBegin    Op = "tx_begin"
	OpTxCommit   Op = "tx_commit"
	OpTxRollback Op = "tx_rollback"
)

// HookInfo contains detailed information about a database operation.
type HookInfo struct {
	Op    Op
	SQL   string
	Args  []any
	InTx  bool
	Start time.Time
	End   time.Time
	Err   error
	Rows  *int64
}

// DBHook defines the interface for database hooks.
type DBHook interface {
	Before(ctx context.Context, info *HookInfo) context.Context
	After(ctx context.Context, info *HookInfo)
}

const (
	ObservabilityLogAll         ObservabilityLogMode = "all"
	ObservabilityLogSlow        ObservabilityLogMode = "slow"
	ObservabilityLogError       ObservabilityLogMode = "error"
	ObservabilityLogSlowOrError ObservabilityLogMode = "slow_or_error"
)

const (
	defaultObservabilitySlowThreshold = 500 * time.Millisecond
	defaultLogFieldMaxSize            = 4 << 10
)

type ObservabilityLogMode string

type DebugHook struct {
	WithArgs bool
}

func (h *DebugHook) Before(ctx context.Context, info *HookInfo) context.Context {
	info.Start = time.Now()
	return ctx
}

func (h *DebugHook) After(ctx context.Context, info *HookInfo) {
	info.End = time.Now()
	dur := info.End.Sub(info.Start)

	attrs := []slog.Attr{
		slog.String("op", string(info.Op)),
		slog.Bool("in_tx", info.InTx),
		slog.Duration("duration", dur),
		slog.Any("err", truncateError(info.Err, defaultLogFieldMaxSize)),
		slog.Any("rows", rowsPtrVal(info.Rows)),
		slog.String("sql", truncateString(info.SQL, defaultLogFieldMaxSize)),
	}

	if h.WithArgs {
		attrs = append(attrs, slog.Any("args", truncateArgs(info.Args, defaultLogFieldMaxSize)))
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "[PGX]", attrs...)
}

type ObservabilityHook struct {
	WithArgs      bool
	Mode          ObservabilityLogMode
	SlowThreshold time.Duration
}

func (h *ObservabilityHook) Before(ctx context.Context, info *HookInfo) context.Context {
	info.Start = time.Now()
	return ctx
}

func (h *ObservabilityHook) After(ctx context.Context, info *HookInfo) {
	info.End = time.Now()
	dur := info.End.Sub(info.Start)
	isSlow := dur >= h.slowThreshold()
	hasErr := info.Err != nil
	if !h.shouldLog(isSlow, hasErr) {
		return
	}

	e := observability.Start(ctx, h.level(isSlow, hasErr)).
		Str("op", string(info.Op)).
		Bool("in_tx", info.InTx).
		Dur("duration", dur).
		Str("sql", truncateString(info.SQL, defaultLogFieldMaxSize))

	if isSlow {
		e = e.Bool("slow", true).
			Dur("slow_threshold", h.slowThreshold())
	}
	if hasErr {
		e = e.Str("err", truncateString(info.Err.Error(), defaultLogFieldMaxSize))
	}
	if info.Rows != nil {
		e = e.Int64("rows", *info.Rows)
	}
	if h.WithArgs {
		e = e.Interface("args", truncateArgs(info.Args, defaultLogFieldMaxSize))
	}

	e.Msg("[PGX]")
}

func (h *ObservabilityHook) shouldLog(isSlow, hasErr bool) bool {
	switch h.logMode() {
	case ObservabilityLogSlow:
		return isSlow
	case ObservabilityLogError:
		return hasErr
	case ObservabilityLogSlowOrError:
		return isSlow || hasErr
	default:
		return true
	}
}

func (h *ObservabilityHook) level(isSlow, hasErr bool) zerolog.Level {
	switch {
	case hasErr:
		return zerolog.ErrorLevel
	case isSlow:
		return zerolog.WarnLevel
	default:
		return zerolog.InfoLevel
	}
}

func (h *ObservabilityHook) logMode() ObservabilityLogMode {
	switch h.Mode {
	case ObservabilityLogSlow, ObservabilityLogError, ObservabilityLogSlowOrError:
		return h.Mode
	default:
		return ObservabilityLogAll
	}
}

func (h *ObservabilityHook) slowThreshold() time.Duration {
	if h.SlowThreshold > 0 {
		return h.SlowThreshold
	}
	return defaultObservabilitySlowThreshold
}

func truncateString(value string, maxSize int) string {
	if maxSize <= 0 || len(value) <= maxSize {
		return value
	}
	return value[:maxSize] + fmt.Sprintf("... [truncated, total size: %d bytes]", len(value))
}

func truncateArgs(args []any, maxSize int) []any {
	if len(args) == 0 {
		return nil
	}

	truncated := make([]any, len(args))
	for i, arg := range args {
		truncated[i] = truncateValue(arg, maxSize)
	}
	return truncated
}

func truncateValue(value any, maxSize int) any {
	switch typed := value.(type) {
	case string:
		return truncateString(typed, maxSize)
	case []byte:
		return truncateString(string(typed), maxSize)
	case fmt.Stringer:
		return truncateString(typed.String(), maxSize)
	default:
		return value
	}
}

func truncateError(err error, maxSize int) any {
	if err == nil {
		return nil
	}
	return truncateString(err.Error(), maxSize)
}

func rowsPtrVal(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}
