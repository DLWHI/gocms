package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type ReasonedHandler struct {
	opts  slog.HandlerOptions
	attrs []slog.Attr
	group string
	mu    *sync.Mutex
}

func NewReasonedHandler(opts *slog.HandlerOptions) *ReasonedHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ReasonedHandler{
		opts: *opts,
		mu:   &sync.Mutex{},
	}
}

func fetchKeyValue(group string, pending *[]slog.Attr) (string, *string) {
	stack := *pending
	current := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	var strval *string
	value := current.Value.Resolve()
	if value.Kind() == slog.KindGroup {
		for _, child := range value.Group() {
			stack = append(stack, slog.Attr{
				Key:   current.Key + "." + child.Key,
				Value: child.Value.Resolve(),
			})
		}
	} else {
		str := fmt.Sprint(value.Any())
		strval = &str
	}

	*pending = stack
	if group != "" {
		return group + "." + current.Key, strval
	} else {
		return current.Key, strval
	}
}

func (h *ReasonedHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *ReasonedHandler) Handle(_ context.Context, r slog.Record) error {
	var line string
	var values string
	reason := ""
	source := "No source"

	tree := make([]slog.Attr, 0, r.NumAttrs()+len(h.attrs))
	r.Attrs(func(a slog.Attr) bool {
		tree = append(tree, a)
		return true
	})

	tree = append(tree, h.attrs...)
	for len(tree) > 0 {
		key, value := fetchKeyValue(h.group, &tree)
		if value != nil {
			switch key {
			case "reason":
				reason = *value
			case "source":
				source = *value
			default:
				values = fmt.Sprintf("%s %s=%s", values, key, safeText(*value))
			}
		}
	}

	timestamp := r.Time.Format(time.DateTime)
	line = fmt.Sprintf("%s [%s] [%s] %s.",
		timestamp,
		r.Level,
		source,
		safeText(r.Message),
	)
	if reason != "" {
		line += fmt.Sprintf(" Reason: %s.", reason)
	}
	if values != "" {
		line += fmt.Sprintf(" Data: %s", values)
	}
	line += "\n"

	return h.write(r.Level, line)
}

func (h *ReasonedHandler) write(level slog.Level, line string) error {
	var err error
	h.mu.Lock()
	defer h.mu.Unlock()
	if level >= slog.LevelWarn {
		_, err = io.WriteString(os.Stderr, line)
	} else {
		_, err = io.WriteString(os.Stdout, line)
	}
	return err
}

func safeText(value string) string {
	value = strings.ReplaceAll(value, "\r", `\r`)
	return strings.ReplaceAll(value, "\n", `\n`)
}

func (h *ReasonedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(h2.attrs, h.attrs)
	if h.group == "" {
		h2.attrs = append(h2.attrs, attrs...)
	} else if len(attrs) > 0 {
		h2.attrs = append(h2.attrs, slog.Attr{
			Key:   h.group,
			Value: slog.GroupValue(attrs...),
		})
	}
	return &h2
}

func (h *ReasonedHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	if name == "" {
		return &h2
	}
	if h2.group != "" {
		h2.group = h2.group + "." + name
	} else {
		h2.group = name
	}
	return &h2
}
