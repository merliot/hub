//go:build !rpi && !x86_64

package camera

import (
	"fmt"
	"io/fs"
	"os"
	"time"
)

var (
	maxMemoryFiles uint32 = 50
	maxFiles       uint32 = 500
)

func startCapture(fileSpec string) error {
	src := "images/0.jpg"
	dst := fmt.Sprintf(fileSpec, 1)

	data, err := fs.ReadFile(embedFS, src)
	if err == nil {
		go func() {
			for {
				os.WriteFile(dst, data, 0644)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	return err
}
