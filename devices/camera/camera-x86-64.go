//go:build x86_64

package camera

import (
	"os/exec"
)

var (
	maxMemoryFiles uint32 = 200
	maxFiles       uint32 = 2000
)

func startCapture(fileSpec string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", "/dev/video0",
		"-vf", "fps=1/5", // Capture every 5 seconds
		"-s", "640x480",
		fileSpec,
	)
	return cmd.Start()
}
