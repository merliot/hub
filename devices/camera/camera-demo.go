//go:build !rpi && !x86_64

package camera

import (
	"fmt"
	"io/fs"
)

func captureJpeg() ([]byte, error) {
	index := 0
	filename := fmt.Sprintf("images/%d.jpg", index)
	data, err := fs.ReadFile(embedFS, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return data, err
}
