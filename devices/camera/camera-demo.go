package camera

import (
	"fmt"
	"io/fs"

	"github.com/merliot/hub/pkg/device"
)

func (c *camera) demoJpeg(index int) ([]byte, error) {
	filename := fmt.Sprintf("images/%d.jpeg", index)
	data, err := fs.ReadFile(embedFS, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return data, err
}

func (c *camera) DemoSetup() error {
	c.jpeg = c.demoJpeg
	return nil
}

func (c *camera) DemoPoll(pkt *device.Packet) {}
