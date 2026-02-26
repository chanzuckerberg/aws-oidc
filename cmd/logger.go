package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

func getDefaultLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/aws-oidc.log"
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	hostname = sanitizeFilename(hostname)

	return filepath.Join(
		homeDir, ".aws-oidc", "logs",
		fmt.Sprintf("aws-oidc.%s.log", hostname),
	)
}

func initLogger(verbosity int, logFile string) (func() error, error) {
	// Default: WARN, -v: INFO, -vv: DEBUG
	stderrLevel := slog.LevelWarn
	switch {
	case verbosity >= 2:
		stderrLevel = slog.LevelDebug
	case verbosity == 1:
		stderrLevel = slog.LevelInfo
	}

	if logFile == "" {
		logFile = getDefaultLogFile()
	}

	logDir := filepath.Dir(logFile)
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}
	closer := func() error {
		return f.Close()
	}

	// Stderr handler respects the verbose flag
	stderrHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: stderrLevel,
	})

	// File handler always logs at DEBUG level
	fileHandler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(&multiHandler{
		handlers: []slog.Handler{stderrHandler, fileHandler},
	})
	slog.SetDefault(logger)
	return closer, nil
}

func sanitizeFilename(s string) string {
	s = filepath.Base(s)
	s = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == '\x00' {
			return '-'
		}
		return r
	}, s)
	if s == "" || s == "." || s == ".." {
		return "unknown"
	}
	return strings.ToLower(s)
}
