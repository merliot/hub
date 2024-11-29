package camera

import (
	"embed"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var embedFS embed.FS

type camera struct {
	jpeg func(int) ([]byte, error)
}

type msgGetImage struct {
	Index int
}

type msgImage struct {
	Jpeg []byte
}

func NewModel() device.Devicer {
	return &camera{}
}

func (c *camera) GetConfig() device.Config {
	return device.Config{
		Model:   "camera",
		State:   c,
		FS:      &embedFS,
		Targets: []string{"rpi"},
		BgColor: "almond-creme",
		FgColor: "black",
		PacketHandlers: device.PacketHandlers{
			"/image": &device.PacketHandler[msgGetImage]{device.RouteUp},
		},
	}
}

func (c *camera) getJpeg(index int) ([]byte, error) {
	return []byte{}, nil
}

func (c *camera) Setup() error {
	c.jpeg = c.getJpeg
	return nil
}

func (c *camera) Poll(pkt *device.Packet) {}
