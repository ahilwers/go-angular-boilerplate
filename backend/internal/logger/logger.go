package logger

import (
	"boilerplate/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func New(cfg config.LoggingConfig) *slog.Logger {
	level := parseLevel(cfg.Level)

	var handler slog.Handler

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	// If Loki is configured, wrap with Loki handler
	if cfg.LokiConfig != nil && cfg.LokiConfig.URL != "" {
		handler = NewLokiHandler(handler, cfg.LokiConfig)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type LokiHandler struct {
	base   slog.Handler
	config *config.LokiConfig
	client *http.Client
}

func NewLokiHandler(base slog.Handler, cfg *config.LokiConfig) *LokiHandler {
	return &LokiHandler{
		base:   base,
		config: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *LokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *LokiHandler) Handle(ctx context.Context, r slog.Record) error {
	// First, let the base handler handle the record (for local logging)
	if err := h.base.Handle(ctx, r); err != nil {
		return err
	}

	// Then, send to Loki asynchronously (don't block on errors)
	go h.sendToLoki(r)

	return nil
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LokiHandler{
		base:   h.base.WithAttrs(attrs),
		config: h.config,
		client: h.client,
	}
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return &LokiHandler{
		base:   h.base.WithGroup(name),
		config: h.config,
		client: h.client,
	}
}

func (h *LokiHandler) sendToLoki(r slog.Record) {
	labels := map[string]string{
		"level": r.Level.String(),
		"app":   "boilerplate",
	}

	attrs := make(map[string]interface{})
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	logLine := map[string]interface{}{
		"message": r.Message,
		"attrs":   attrs,
	}

	logLineJSON, err := json.Marshal(logLine)
	if err != nil {
		return
	}

	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": labels,
				"values": [][]string{
					{
						fmt.Sprintf("%d", r.Time.UnixNano()),
						string(logLineJSON),
					},
				},
			},
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", h.config.URL+"/loki/api/v1/push", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if h.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.config.BearerToken)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
}
