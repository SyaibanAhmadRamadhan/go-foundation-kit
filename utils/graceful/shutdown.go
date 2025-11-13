package graceful

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Shutdown(fn func(ctx context.Context) error, timeout time.Duration, osSignals ...os.Signal) {
	quit := make(chan os.Signal, 1)

	if len(osSignals) == 0 {
		osSignals = append(osSignals, syscall.SIGINT, syscall.SIGTERM)
	}
	signal.Notify(quit, osSignals...)

	sig := <-quit
	slog.Info("Received OS signal, starting graceful shutdown", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := fn(ctx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	} else {
		slog.Info("Graceful shutdown completed successfully")
	}
}

func TriggerShutdown() {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		slog.Error("Failed to find process", "error", err)
		return
	}

	if err := p.Signal(syscall.SIGINT); err != nil {
		slog.Error("Failed to send shutdown signal", "error", err)
	}
}
