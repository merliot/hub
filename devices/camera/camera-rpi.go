//go:build rpi

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
		"libcamera-still",
		"-t", "0", // Continuous capture
		"--timelapse", "5s", // Capture every 5 seconds
		"--width", "640",
		"--height", "480",
		"-o", fileSpec,
	)
	return cmd.Start()
}
