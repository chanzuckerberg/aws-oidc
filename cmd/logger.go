package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/loki/client"
	"github.com/chanzuckerberg/aws-oidc/pkg/util"
)

// multiHandler fans out log records to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			err := handler.Handle(ctx, record)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// lokiBuffer holds entries to push; shared by handler copies (WithAttrs/WithGroup).
type lokiBuffer struct {
	mu      sync.Mutex
	entries []client.Entry
}

// lokiHandler buffers log records and pushes them to Loki on Flush via pkg/loki/client.
type lokiHandler struct {
	opts        *slog.HandlerOptions
	client      *client.Client
	buf         *lokiBuffer
	labels      map[string]string
	jsonHandler slog.Handler
}

func newLokiHandler(pushURL, hostname, username, password string, level slog.Level) (*lokiHandler, error) {
	userAgent := "aws-oidc"
	if v, err := util.VersionString(); err == nil && v != "" {
		userAgent = "aws-oidc/" + v
	}
	c := client.NewClient(pushURL, userAgent, username, password, 10*time.Second, 5)
	labels := map[string]string{
		"job":      "aws-oidc",
		"hostname": hostname,
	}
	opts := &slog.HandlerOptions{Level: level}
	return &lokiHandler{
		opts:        opts,
		client:      c,
		buf:         &lokiBuffer{},
		labels:      labels,
		jsonHandler: slog.NewJSONHandler(io.Discard, opts),
	}, nil
}

func (h *lokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *lokiHandler) Handle(ctx context.Context, record slog.Record) error {
	var buf bytes.Buffer
	th := slog.NewJSONHandler(&buf, h.opts)
	err := th.Handle(ctx, record)
	if err != nil {
		return err
	}
	line := strings.TrimSuffix(buf.String(), "\n")
	h.buf.mu.Lock()
	h.buf.entries = append(h.buf.entries, client.Entry{Timestamp: record.Time, Line: line})
	h.buf.mu.Unlock()
	return nil
}

func (h *lokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &lokiHandler{
		opts:        h.opts,
		client:      h.client,
		buf:         h.buf,
		labels:      h.labels,
		jsonHandler: h.jsonHandler.WithAttrs(attrs),
	}
}

func (h *lokiHandler) WithGroup(name string) slog.Handler {
	return &lokiHandler{
		opts:        h.opts,
		client:      h.client,
		buf:         h.buf,
		labels:      h.labels,
		jsonHandler: h.jsonHandler.WithGroup(name),
	}
}

func (h *lokiHandler) Flush() error {
	h.buf.mu.Lock()
	entries := h.buf.entries
	h.buf.entries = nil
	h.buf.mu.Unlock()
	if len(entries) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := h.client.Push(ctx, h.labels, entries)
	if err != nil {
		return fmt.Errorf("pushing logs to Loki: %w", err)
	}
	return nil
}

// parseBasicAuthCredentials decodes a base64-encoded "username:password" and returns the two parts.
func parseBasicAuthCredentials(base64Value string) (username, password string, err error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Value)
	if err != nil {
		return "", "", fmt.Errorf("credentials must be base64-encoded username:password: %w", err)
	}
	line := string(decoded)
	idx := strings.Index(line, ":")
	if idx == -1 {
		return "", "", fmt.Errorf("credentials must decode to username:password")
	}
	username = strings.TrimSpace(line[:idx])
	password = strings.TrimSpace(line[idx+1:])
	if username == "" || password == "" {
		return "", "", fmt.Errorf("username and password must be non-empty")
	}
	return username, password, nil
}

func initLogger(verbosity int, logLokiURL, logLokiCredentials string) (func() error, error) {
	// Default: WARN, -v: INFO, -vv: DEBUG
	stderrLevel := slog.LevelWarn
	switch {
	case verbosity >= 2:
		stderrLevel = slog.LevelDebug
	case verbosity == 1:
		stderrLevel = slog.LevelInfo
	}

	stderrHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: stderrLevel,
	})

	handlers := []slog.Handler{stderrHandler}

	closer := func() error { return nil }
	var loki *lokiHandler
	username, password, err := parseBasicAuthCredentials(logLokiCredentials)
	if err != nil {
		return closer, fmt.Errorf("Loki credentials: %w", err)
	}
	if logLokiURL != "" && username != "" && password != "" {
		baseURL := strings.TrimSuffix(logLokiURL, "/")
		hostname, _ := os.Hostname()
		var err error
		loki, err = newLokiHandler(baseURL, hostname, username, password, slog.LevelDebug)
		if err != nil {
			return closer, fmt.Errorf("initializing Loki handler: %w", err)
		}
		handlers = append(handlers, loki)
	}

	logger := slog.New(&multiHandler{handlers: handlers})
	slog.SetDefault(logger)

	closer = func() error {
		if loki != nil {
			return loki.Flush()
		}
		return nil
	}
	return closer, nil
}
