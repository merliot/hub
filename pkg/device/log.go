// Poor-man's version of slog.  slog wasn't working for me on tinygo...kept
// running OOM.

package device

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

var (
	logBuffer   bytes.Buffer
	logBufferMu mutex
)

// Format the args into key=value pairs
func formatArgs(args ...any) string {
	logBufferMu.Lock()
	defer logBufferMu.Unlock()

	logBuffer.Reset()

	for i := 0; i < len(args); i += 2 {
		switch {
		case i == 0:
			logBuffer.WriteString(" ")
		case i > 0:
			logBuffer.WriteString(", ")
		}
		fmt.Fprintf(&logBuffer, "%v=%v", args[i], args[i+1])
	}
	return logBuffer.String()
}

// ANSI escape codes for colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
)

func getColor(level string) string {
	switch level {
	case "ERROR":
		return colorRed
	case "WARN":
		return colorYellow
	case "INFO":
		return colorGreen
	case "DEBUG":
		return colorBlue
	default:
		return colorReset
	}
}

func (s *server) logok(level string) bool {
	switch s.logLevel {
	case "DEBUG":
		return true
	case "INFO":
		return level == "INFO" || level == "WARN" || level == "ERROR"
	case "WARN":
		return level == "WARN" || level == "ERROR"
	case "ERROR":
		return level == "ERROR"
	}
	return false
}

func (s *server) log(level string, msg string, args ...any) {
	if s.logok(level) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		color := getColor(level)
		port := ""
		if s.port != 0 {
			port = "[:" + strconv.Itoa(s.port) + "]"
		}
		fmt.Printf("%s %s%s[%s]%s %s%s%s", timestamp, port, color,
			level, colorReset, msg, formatArgs(args...), crlf)
	}
}

func (s *server) logInfo(msg string, args ...any) {
	s.log("INFO", msg, args...)
}

func (s *server) logWarn(msg string, args ...any) {
	s.log("WARN", msg, args...)
}

func (s *server) logDebug(msg string, args ...any) {
	s.log("DEBUG", msg, args...)
}

func (s *server) logError(msg string, args ...any) {
	s.log("ERROR", msg, args...)
}
