package slogpretty

import (
	"context"
	"encoding/json"
	"io"
	stdLog "log"
	"log/slog"

	"github.com/fatih/color"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}

	return h
}

func addAttrs(attrs []slog.Attr, target map[string]any) {
	for _, a := range attrs {
		if group, ok := a.Value.Any().([]slog.Attr); ok {
			groupMap := make(map[string]any)
			addAttrs(group, groupMap)
			target[a.Key] = groupMap
		} else {
			target[a.Key] = a.Value.Any()
		}
	}
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]any, r.NumAttrs())

	// Добавляем атрибуты из записи
	r.Attrs(func(a slog.Attr) bool {
		if group, ok := a.Value.Any().([]slog.Attr); ok {
			groupMap := make(map[string]any)
			addAttrs(group, groupMap)
			fields[a.Key] = groupMap
		} else {
			fields[a.Key] = a.Value.Any()
		}
		return true
	})

	// Добавляем атрибуты из хендлера
	addAttrs(h.attrs, fields)

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	timeStr := r.Time.Format("[15:04:05.000]")
	msg := color.CyanString(r.Message)

	h.l.Println(
		color.HiWhiteString(timeStr),
		level,
		msg,
		string(b),
	)

	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
		attrs:   append(h.attrs, slog.Group(name)),
	}
}
