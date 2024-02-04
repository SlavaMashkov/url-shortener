package slogdiscard

import (
	"context"
	"log/slog"
)

func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

type DiscardHandler struct{}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	// Ignore logging
	return nil
}

func (h DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// Return self
	return h
}

func (h DiscardHandler) WithGroup(_ string) slog.Handler {
	// Return self
	return h
}

func (h DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// Always returns false
	return false
}
