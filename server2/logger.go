package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"
)

var logger *slog.Logger

type PipeHandler struct {
	w  io.Writer
	ch chan string
}

func NewPipeHandler(w io.Writer, stringchan chan string) *PipeHandler {
	return &PipeHandler{
		w:  w,
		ch: stringchan,
	}
}

func (h *PipeHandler) Handle(ctx context.Context, r slog.Record) error {

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "time=%s level=%s server=2",
		r.Time.Format(time.RFC3339),
		r.Level.String())

	fmt.Fprintf(&buf, " msg=%s", r.Message)

	// Добавляем все атрибуты
	r.Attrs(func(attr slog.Attr) bool {
		fmt.Fprintf(&buf, " %s=%v", attr.Key, attr.Value.Any())
		return true
	})

	buf.WriteByte('\n')

	logLine := buf.String()

	h.ch <- logLine
	_, err := io.WriteString(h.w, logLine)
	return err
}

func (h *PipeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *PipeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *PipeHandler) WithGroup(name string) slog.Handler {
	return h
}
