/*
Copyright © 2024 The Dues Authors
*/
package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

const (
	timeFormat = "[15:04:05.000]"
)

type Handler struct {
	handler slog.Handler
	buffer  *bytes.Buffer
	mutex   *sync.Mutex
	writer  io.Writer
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{handler: h.handler.WithAttrs(attrs), buffer: h.buffer, mutex: h.mutex}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{handler: h.handler.WithGroup(name), buffer: h.buffer, mutex: h.mutex}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = Colorize(DarkGray, level)
	case slog.LevelInfo:
		level = Colorize(Cyan, level)
	case slog.LevelWarn:
		level = Colorize(LightYellow, level)
	case slog.LevelError:
		level = Colorize(LightRed, level)
	}

	customWriter := DuesWriter{
		writer:   h.writer,
		time:     Colorize(LightGray, r.Time.Format(time.RFC850)),
		command:  Colorize(LightGreen, "root-dues"),
		severity: level,
	}

	attributes, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	if len(attributes) > 0 {

		bytes, err := json.MarshalIndent(attributes, "", "  ")

		if err != nil {
			return fmt.Errorf("error when marshalling attributes: %w", err)
		}

		fmt.Fprintf(&customWriter, "%s\n%s\n", Colorize(White, r.Message), Colorize(DarkGray, string(bytes)))
	} else {

		fmt.Fprintf(&customWriter, "%s\n", Colorize(White, r.Message))
	}

	return nil
}

func (h *Handler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.mutex.Lock()

	defer func() {
		h.buffer.Reset()
		h.mutex.Unlock()
	}()

	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("Error occured while computing attributes for internal custom writer handler")
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buffer.Bytes(), &attrs)

	if err != nil {
		return nil, fmt.Errorf("error when marshalling attributes for internal custom writer handler")
	}

	return attrs, nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,

) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

func NewDuesHandler(writer io.Writer, opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	if writer == nil {
		writer = os.Stdout
	}
	buffer := &bytes.Buffer{}
	return &Handler{
		buffer: buffer,
		handler: slog.NewJSONHandler(buffer, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		mutex:  &sync.Mutex{},
		writer: writer,
	}
}
