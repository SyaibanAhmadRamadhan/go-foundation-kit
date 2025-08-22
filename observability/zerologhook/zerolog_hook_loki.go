package zerologhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

// LokiPayload represents the payload structure required by Loki's push API.
type LokiPayload struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream holds a single stream's labels and the batched log entries.
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [timestamp, log message]
}

// LokiHookConfig configures the Loki log hook.
type LokiHookConfig struct {
	Username          string        // Basic Auth username
	Password          string        // Basic Auth password
	Endpoint          string        // Loki push endpoint
	Env               string        // Environment name (e.g., production)
	ServiceName       string        // Name of the service
	OnlySink          bool          // Flag if this hook is only used as sink
	BatchInterval     time.Duration // Time interval for flushing logs
	BatchMessageCount int           // Max number of logs per batch
}

// lokiHook buffers log entries and pushes them to Loki in batches.
type lokiHook struct {
	username    string
	password    string
	endpoint    string
	env         string
	serviceName string
	onlySink    bool

	batchMessageCount int
	mu                sync.Mutex
	lokiStream        []LokiStream
	ticker            *time.Ticker
	stopChan          chan struct{}
	wg                sync.WaitGroup
}

// NewLokiHook initializes the log hook and starts the background batch flusher.
// Returns the hook and a shutdown function.
func NewLokiHook(cfg LokiHookConfig) (*lokiHook, func()) {
	hook := &lokiHook{
		username:    cfg.Username,
		password:    cfg.Password,
		endpoint:    cfg.Endpoint,
		env:         cfg.Env,
		serviceName: cfg.ServiceName,
		onlySink:    cfg.OnlySink,

		batchMessageCount: cfg.BatchMessageCount,
		lokiStream:        make([]LokiStream, 0, cfg.BatchMessageCount),
		stopChan:          make(chan struct{}),
		ticker:            time.NewTicker(cfg.BatchInterval),
	}

	// Start background sender
	hook.wg.Add(1)
	go hook.batchSender()

	return hook, func() {
		slog.Info("shutting down loki hook...",
			slog.String("env", hook.env),
			slog.String("service", hook.serviceName),
		)
		close(hook.stopChan)
		hook.wg.Wait()
		hook.ticker.Stop()
	}
}

// Write adds a log message to the batch. If batch is full, triggers flush.
func (w *lokiHook) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var payload map[string]any
	if err := json.Unmarshal(p, &payload); err != nil {
		slog.Error("LokiLogWriter: failed to parse log JSON",
			slog.String("service", w.serviceName),
			slog.String("env", w.env),
			slog.Any("error", err),
			slog.String("raw", string(p)),
		)
		return len(p), nil
	}

	level := payload["level"]
	statusCode := payload["status_code"]
	spanID := payload["span_id"]
	traceID := payload["trace_id"]
	stream := map[string]string{
		"app":   os.Getenv("APP_NAME"),
		"env":   os.Getenv("APP_ENV"),
		"level": fmt.Sprintf("%v", level),
	}
	if statusCode != nil {
		stream["status_code"] = fmt.Sprintf("%v", statusCode)
	}
	if spanID != nil {
		stream["span_id"] = fmt.Sprintf("%v", spanID)
	}
	if traceID != nil {
		stream["trace_id"] = fmt.Sprintf("%v", traceID)
	}

	if traceID == nil {
		slog.Warn("LokiLogWriter: log entry missing trace_id",
			slog.String("service", w.serviceName),
			slog.String("env", w.env),
			slog.Any("payload", payload),
		)
	}

	now := time.Now().UnixNano()
	lokiStream := LokiStream{
		Stream: stream,
		Values: [][2]string{
			{fmt.Sprintf("%d", now), string(p)},
		},
	}

	w.lokiStream = append(w.lokiStream, lokiStream)

	if len(w.lokiStream) > w.batchMessageCount {
		w.flush()
	}

	return
}

// batchSender runs in background to flush logs periodically or on shutdown.
func (w *lokiHook) batchSender() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ticker.C:
			w.mu.Lock()
			w.flush()
			w.mu.Unlock()
		case <-w.stopChan:
			w.mu.Lock()
			w.flush()
			w.mu.Unlock()
			return
		}
	}
}

// flush sends the current batch to Loki using HTTP POST with Basic Auth.
func (w *lokiHook) flush() {
	if len(w.lokiStream) == 0 {
		return
	}

	bodyReqStruct := LokiPayload{
		Streams: w.lokiStream,
	}

	body, err := json.Marshal(bodyReqStruct)
	if err != nil {
		slog.Error("LokiLogWriter: failed to marshal body loki JSON",
			slog.String("service", w.serviceName),
			slog.String("env", w.env),
			slog.Any("error", err),
			slog.Any("raw", bodyReqStruct),
		)
		return
	}

	req, err := http.NewRequest(http.MethodPost, w.endpoint, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("LokiLogWriter: failed to create new http request loki",
			slog.String("service", w.serviceName),
			slog.String("env", w.env),
			slog.Any("error", err),
			slog.String("raw", string(body)),
		)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(w.username, w.password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("LokiLogWriter: failed to send http request loki",
			slog.String("service", w.serviceName),
			slog.String("env", w.env),
			slog.Any("error", err),
			slog.String("raw", string(body)),
		)
		return
	}
	defer resp.Body.Close()

	w.lokiStream = w.lokiStream[:0]
}
