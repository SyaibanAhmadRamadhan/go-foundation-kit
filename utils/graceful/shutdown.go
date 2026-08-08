package graceful

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type CloseFunc func(context.Context) error

type closeResource struct {
	name string
	fn   CloseFunc
}

type config struct {
	timeout time.Duration
	signals []os.Signal
}

type Option func(*config)

var (
	mu sync.Mutex

	cfg = config{
		timeout: 30 * time.Second,
		signals: []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
		},
	}

	resources []closeResource

	listenerOnce sync.Once
	shutdownOnce sync.Once

	started bool
)

func Configure(options ...Option) error {
	mu.Lock()
	defer mu.Unlock()

	if started {
		return errors.New("graceful: already started")
	}

	for _, option := range options {
		option(&cfg)
	}

	return nil
}

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *config) {
		cfg.timeout = timeout
	}
}

func WithSignals(signals ...os.Signal) Option {
	return func(cfg *config) {
		cfg.signals = append([]os.Signal(nil), signals...)
	}
}

func RegisterCloseResource(
	name string,
	fn CloseFunc,
) {
	mu.Lock()

	resources = append(resources, closeResource{
		name: name,
		fn:   fn,
	})

	mu.Unlock()

	startListener()
}

func startListener() {
	listenerOnce.Do(func() {
		mu.Lock()

		started = true

		timeout := cfg.timeout
		signals := append([]os.Signal(nil), cfg.signals...)

		mu.Unlock()

		ctx, stop := signal.NotifyContext(
			context.Background(),
			signals...,
		)

		go func() {
			defer stop()

			<-ctx.Done()

			slog.Info(
				"Received OS signal, starting graceful shutdown",
			)

			shutdown(timeout)
		}()
	})
}

func shutdown(timeout time.Duration) {
	shutdownOnce.Do(func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			timeout,
		)
		defer cancel()

		if err := closeAll(ctx); err != nil {
			slog.Error(
				"Graceful shutdown completed with errors",
				"error", err,
			)
			return
		}

		slog.Info(
			"Graceful shutdown completed successfully",
		)
	})
}

func closeAll(ctx context.Context) error {
	mu.Lock()
	items := append([]closeResource(nil), resources...)
	mu.Unlock()

	var errs []error

	for i := len(items) - 1; i >= 0; i-- {
		resource := items[i]

		slog.Info(
			"Closing resource",
			"resource", resource.name,
		)

		if err := resource.fn(ctx); err != nil {
			errs = append(
				errs,
				fmt.Errorf("%s: %w", resource.name, err),
			)

			slog.Error(
				"Failed to close resource",
				"resource", resource.name,
				"error", err,
			)

			continue
		}

		slog.Info(
			"Resource closed successfully",
			"resource", resource.name,
		)
	}

	return errors.Join(errs...)
}

func TriggerShutdown() {
	mu.Lock()
	timeout := cfg.timeout
	mu.Unlock()

	shutdown(timeout)
}
