//go:build tinygo

package hub

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

var logLevel = slog.LevelInfo

// Custom writer to replace \n with \r\n
type crlfWriter struct {
	io.Writer
}

func (cw *crlfWriter) Write(p []byte) (n int, err error) {
	// Replace all occurrences of \n with \r\n
	crlf := []byte(fmt.Sprintf("%s\r\n", p[:len(p)-1])) // Skip last \n and add \r\n
	return cw.Write(crlf)
}

func setupLogger() {
	// wait a bit for serial
	time.Sleep(2 * time.Second)

	writer := &crlfWriter{os.Stdout}
	handler := newLogHandler(writer)
	slog.SetDefault(slog.New(handler))
}
