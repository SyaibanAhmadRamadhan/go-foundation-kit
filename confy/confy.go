package confy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/knadh/koanf/v2"
)

// Loader is the main config loader.
// It loads config from file(s), merges env variables,
// unmarshals into a struct, and supports watching for file changes.
type Loader[T any] struct {
	k   *koanf.Koanf      // underlying koanf instance
	cur atomic.Pointer[T] // holds the current config snapshot
	opt options           // applied options

	ctx    context.Context
	cancel context.CancelFunc
}

func New[T any](opts ...Option) (*Loader[T], error) {
	o := options{
		delimiter: ".",
		tag:       "koanf",
	}

	k := koanf.New(o.delimiter)

	ctx, cancel := context.WithCancel(context.Background())
	l := &Loader[T]{
		k:      k,
		opt:    o,
		cur:    atomic.Pointer[T]{},
		ctx:    ctx,
		cancel: cancel,
	}

	if len(l.opt.providers) == 0 {
		return nil, errors.New("confy: no providers; use SetProvider or SetProviders")
	}

	// Initial load
	if err := l.loadOnce(); err != nil {
		return nil, err
	}

	l.startIntervalRefresh()

	return l, nil
}

func (l *Loader[T]) loadOnce() error {
	var errs error
	for _, p := range l.opt.providers {
		if err := l.k.Load(p, p.Parser()); err != nil {
			errs = errors.Join(errs, fmt.Errorf("confy: loading provider %s: %w", p.Name(), err))
		} else {
			// Successfully loaded provider
			slog.Info("load config successfully", slog.Attr{
				Key:   "provider_name",
				Value: slog.StringValue(p.Name()),
			})
		}
	}

	if errs != nil {
		return errs
	}

	var cfg T
	if err := l.k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		Tag:       l.opt.tag,
		FlatPaths: l.opt.flat,
	}); err != nil {
		return err
	}
	l.cur.Store(&cfg)
	return nil
}

func (l *Loader[T]) Get() *T {
	return l.cur.Load()
}

func (l *Loader[T]) ReloadAndGet() (*T, error) {
	if err := l.loadOnce(); err != nil {
		return nil, err
	}
	return l.cur.Load(), nil
}

func (l *Loader[T]) Close() {
	l.cancel()
}

func (l *Loader[T]) startIntervalRefresh() {
	if l.opt.refreshInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(l.opt.refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := l.loadOnce(); err != nil {
					slog.Error("confy: interval refresh: load config", slog.String("error", err.Error()))
					continue
				}
				slog.Info("confy: interval refresh: config reloaded successfully")

				if l.opt.onRefresh != nil {
					l.opt.onRefresh(l.k.Raw())
				}

				Notify()
			case <-l.ctx.Done():
				slog.Info("confy: interval refresh: stopped")
				return
			}
		}
	}()
}
