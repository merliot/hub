//go:build rpi

package camera

import (
	"fmt"
	"os/exec"
	"time"
)

func (c *camera) poll() {

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("image_%s.jpg", timestamp)

	cmd := exec.Command("libcamera-still", "-o", filename, "--width", "640", "--height", "480", "-t", "1", "--immediate")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error capturing image: %v\n", err)
	} else {
		fmt.Printf("Captured %s\n", filename)
		c.Lock()
		c.latest = filename
		c.Unlock()
	}
}
