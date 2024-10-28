package hub

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// ANSI escape codes for colors
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
)

// logHandler implements slog.Handler and colorizes log entries based on level.
type logHandler struct {
	io.Writer
}

func newLogHandler(w io.Writer) *logHandler {
	return &logHandler{w}
}

// getColor returns the appropriate color for a given log level.
func getColor(level slog.Level) string {
	switch level {
	case slog.LevelError:
		return Red
	case slog.LevelWarn:
		return Yellow
	case slog.LevelInfo:
		return Reset
	case slog.LevelDebug:
		return Blue
	default:
		return Reset
	}
}

// formatAttrs collects and formats attributes from the log record.
func formatAttrs(rec slog.Record) string {
	var sb strings.Builder
	rec.Attrs(func(attr slog.Attr) bool {
		fmt.Fprintf(&sb, "%s=%v ", attr.Key, attr.Value)
		return true // Continue iteration
	})
	return strings.TrimSpace(sb.String()) // Remove trailing space
}

// Handle formats and outputs the log record with colorized output.
func (h *logHandler) Handle(ctx context.Context, rec slog.Record) error {
	// Custom time formatting
	timestamp := rec.Time.Format("2006-01-02 15:04:05")

	// Colorized level string
	color := getColor(rec.Level)
	level := rec.Level.String()

	// Collect attributes into a formatted string
	attrStr := formatAttrs(rec)

	// Format log line with time, colorized level, message, and attributes
	logLine := fmt.Sprintf("[%s] %s[%s]%s %s", timestamp, color, level, Reset, rec.Message)
	if attrStr != "" {
		logLine += fmt.Sprintf(" %s", attrStr)
	}
	logLine += "\n"

	// Write the log line to the provided writer (e.g., stdout or file)
	_, err := h.Write([]byte(logLine))
	return err
}

func (h *logHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= logLevel
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return h
}
