package camera

import (
	"fmt"
	"io/fs"

	"github.com/merliot/hub/pkg/device"
)

func (c *camera) demoRawJpeg(index uint) ([]byte, error) {
	filename := fmt.Sprintf("images/%d.jpg", index)
	data, err := fs.ReadFile(embedFS, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return data, err
}

func (c *camera) DemoSetup() error {
	//c.getJpeg = c.demoRawJpeg
	return nil
}

func (c *camera) DemoPoll(pkt *device.Packet) {}
