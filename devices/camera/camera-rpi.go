//go:build rpi

package camera

import (
	"bytes"
	"fmt"
	"os/exec"
)

func captureJpeg() ([]byte, error) {

	// Create a command for libcamera-still to output to stdout
	libcameraCmd := exec.Command("libcamera-still", "-o", "-", "--width", "640",
		"--height", "480", "-t", "1", "--immediate")

	// Create a command for ffmpeg to read from stdin and add a timestamp
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", "pipe:0", // Read input from stdin
		"-vf", "drawtext=text='%{localtime}':fontcolor=black:fontsize=24:x=10:y=10",
		"-q:v", "7", // JPEG quality
		"-f", "image2", // Output format
		"pipe:1", // Write output to stdout
	)

	// Set up a pipe between the two commands
	libcameraStdout, err := libcameraCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for libcamera-still: %v", err)
	}
	ffmpegCmd.Stdin = libcameraStdout

	// Capture FFmpeg's output
	var ffmpegOutput bytes.Buffer
	ffmpegCmd.Stdout = &ffmpegOutput

	// Start libcamera-still
	if err := libcameraCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start libcamera-still: %v", err)
	}

	// Start ffmpeg
	if err := ffmpegCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	// Wait for libcamera-still to finish
	if err := libcameraCmd.Wait(); err != nil {
		return nil, fmt.Errorf("libcamera-still command failed: %v", err)
	}

	// Wait for ffmpeg to finish
	if err := ffmpegCmd.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg command failed: %v", err)
	}

	// Return the processed image data
	return ffmpegOutput.Bytes(), nil
}
