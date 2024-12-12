//go:build !rpi && !x86_64

package camera

import (
	"bytes"
	"fmt"
	"io/fs"
	"os/exec"
)

func captureJpeg() ([]byte, error) {
	index := 0
	filename := fmt.Sprintf("images/%d.jpg", index)
	jpeg, err := fs.ReadFile(embedFS, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	// Prepare the FFmpeg command to read from stdin and write to stdout
	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0", // Input from stdin
		"-vf", "drawtext=text='%{localtime}':x=10:y=10:fontcolor=black:fontsize=24",
		"-f", "image2", // Output format
		"-vcodec", "mjpeg", // Output codec
		"pipe:1", // Output to stdout
	)

	// Set up pipes for stdin and stdout
	var output bytes.Buffer
	cmd.Stdin = bytes.NewReader(jpeg)
	cmd.Stdout = &output

	// Capture errors
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run FFmpeg command: %v, %s", err, stderr.String())
	}

	// The output buffer now contains the processed JPEG
	return output.Bytes(), nil
}
