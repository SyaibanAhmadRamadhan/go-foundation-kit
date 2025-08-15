package sqlx

import (
	"context"
	"log/slog"
	"time"
)

type DebugHook struct {
	WithArgs bool // If true, include SQL args in the log
}

func (h *DebugHook) Before(ctx context.Context, info *HookInfo) context.Context {
	info.Start = time.Now()
	return ctx
}

func (h *DebugHook) After(ctx context.Context, info *HookInfo) {
	info.End = time.Now()
	dur := info.End.Sub(info.Start)

	// Base fields for all logs
	attrs := []slog.Attr{
		slog.String("op", string(info.Op)),
		slog.Bool("in_tx", info.InTx),
		slog.Bool("prepared", info.Prepared),
		slog.Any("cache_hit", boolPtrVal(info.CacheHit)),
		slog.Duration("duration", dur),
		slog.Any("err", info.Err),
		slog.Any("rows", rowsPtrVal(info.Rows)),
		slog.String("sql", info.SQL),
	}

	// Optionally add args
	if h.WithArgs {
		attrs = append(attrs, slog.Any("args", info.Args))
	}

	// Log as info
	slog.LogAttrs(ctx, slog.LevelInfo, "[SQL]", attrs...)
}

func rowsPtrVal(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func boolPtrVal(p *bool) any {
	if p == nil {
		return nil
	}
	return *p
}
