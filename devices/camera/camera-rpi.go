//go:build rpi

package camera

import (
	"bytes"
	"fmt"
	"os/exec"
)

// captureImage captures a jpeg image using libcamera-still
func captureJpeg() ([]byte, error) {

	cmd := exec.Command("libcamera-still", "-o", "-", "--width", "640",
		"--height", "480", "-t", "1", "--immediate")

	// Create a buffer to capture the output (stdout) from libcamera-still
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = nil // Suppress error output, handle separately if needed

	// Execute the command
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("libcamera-still command failed: %v", err)
	}

	// Return the captured image data
	return outBuf.Bytes(), nil
}
