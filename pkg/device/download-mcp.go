//go:build !tinygo

package device

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
)

func (s *server) downloadMCPServer(w http.ResponseWriter, r *http.Request) {
	arch := r.PathValue("arch")

	// Check if binary exists
	binary := fmt.Sprintf("bin/mcp-server-%s", arch)
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Binary %s not found", binary), http.StatusNotFound)
		return
	}

	hubURL := r.Referer()

	// Create a temporary file for the self-extracting script
	tmpFile, err := os.CreateTemp("", "mcp-server-*.sh")
	if err != nil {
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())

	// Write the script header
	scriptHeader := fmt.Sprintf(`#!/bin/bash

# Set environment variables
export HUB_URL="%s"
export USER="%s"
export PASSWD="%s"

# Create a temporary file for the binary
TEMP_BINARY=$(mktemp)

# Find the line number of the EOF marker
EOF_LINE=$(grep -n '^EOF$' "$0" | cut -d: -f1)

# Extract the binary (everything after the EOF marker) using tail
tail -n +$((EOF_LINE + 1)) "$0" > "$TEMP_BINARY"

# Make the temporary file executable
chmod +x "$TEMP_BINARY"

# Run the binary
"$TEMP_BINARY"

# Clean up
rm -f "$TEMP_BINARY"

exit 0

# Marker for the end of the script content
EOF`, hubURL, s.user, s.passwd)

	if _, err := tmpFile.WriteString(scriptHeader); err != nil {
		http.Error(w, "Failed to write script header", http.StatusInternalServerError)
		return
	}

	// Append the binary
	binaryData, err := os.ReadFile(binary)
	if err != nil {
		http.Error(w, "Failed to read binary", http.StatusInternalServerError)
		return
	}

	if _, err := tmpFile.Write(binaryData); err != nil {
		http.Error(w, "Failed to append binary", http.StatusInternalServerError)
		return
	}

	// Close the temp file
	if err := tmpFile.Close(); err != nil {
		http.Error(w, "Failed to close temporary file", http.StatusInternalServerError)
		return
	}

	// Create a gzip buffer
	var gzipBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuf)

	// Read the temp file
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		http.Error(w, "Failed to read temporary file", http.StatusInternalServerError)
		return
	}

	// Write to gzip
	if _, err := gzipWriter.Write(data); err != nil {
		http.Error(w, "Failed to compress data", http.StatusInternalServerError)
		return
	}

	if err := gzipWriter.Close(); err != nil {
		http.Error(w, "Failed to finalize compression", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s-mcp-server-%s.sh.gz"`,
		s.root.Model, s.root.Id, arch))

	// Write the gzipped data
	if _, err := w.Write(gzipBuf.Bytes()); err != nil {
		http.Error(w, "Failed to write gzipped data", http.StatusInternalServerError)
		return
	}
}
