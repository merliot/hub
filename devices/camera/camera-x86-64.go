//go:build x86_64

package camera

import (
	"bytes"
	"fmt"
	"os/exec"
)

// captureImage captures a jpeg image using ffmpeg
func captureJpeg() ([]byte, error) {

	cmd := exec.Command("ffmpeg", "-f", "v4l2", "-framerate", "30",
		"-video_size", "640x480", "-i", "/dev/video0", "-vframes", "1",
		"-vf", "drawtext=text='%{localtime}':fontcolor=black:fontsize=24:x=10:y=10",
		"-q:v", "7", "-f", "image2pipe", "-vcodec", "mjpeg", "-")

	// Create a buffer to capture the output (stdout) from ffmpeg
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = nil // Suppress error output, handle separately if needed

	// Execute the command
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %v", err)
	}

	// Return the captured image data
	return outBuf.Bytes(), nil
}
