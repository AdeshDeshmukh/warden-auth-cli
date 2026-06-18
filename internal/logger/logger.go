package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

func New(logPath string) (*slog.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	jsonHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	textHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	combined := &multiHandler{
		handlers: []slog.Handler{jsonHandler, textHandler},
	}

	return slog.New(combined), nil
}