//go:build !tinygo

package hub

import (
	"log/slog"
	"os"
)

var logLevel slog.Level

func setupLogger() {

	var level = Getenv("LOG_LEVEL", "INFO")
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := newLogHandler(os.Stdout)
	slog.SetDefault(slog.New(handler))
}
