package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/fatih/color"
)

// TODO add mutex on currentColor
var currentColor *color.Color
var logger *slog.Logger

var colors = map[string]*color.Color{
	"red":    color.New(color.FgRed),
	"blue":   color.New(color.FgBlue),
	"yellow": color.New(color.FgYellow),
	"green":  color.New(color.FgGreen),
	"white":  color.New(color.FgWhite),
	"cyan":   color.New(color.FgCyan),
}

type ColorHandler struct {
	w  io.Writer
	ch chan string
}

func NewColorHandler(w io.Writer, stringchan chan string) *ColorHandler {
	return &ColorHandler{
		w:  w,
		ch: stringchan,
	}
}

func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	// Создаем буфер для построения строки лога
	var buf bytes.Buffer

	// Базовые поля
	fmt.Fprintf(&buf, "time=%s level=%s server=1",
		r.Time.Format(time.RFC3339),
		r.Level.String())

	// Основное сообщение с цветом
	fmt.Fprintf(&buf, " msg=%s", currentColor.Sprint(r.Message))

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

func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return h
}

func changeColor(color string) bool {
	if realColor, ok := colors[color]; ok {
		currentColor = realColor
		return true
	}
	return false
}
