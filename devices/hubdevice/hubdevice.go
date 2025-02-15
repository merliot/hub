package hubdevice

import (
	"embed"

	"github.com/merliot/hub/pkg/device"
)

//go:embed images *.go template
var fs embed.FS

type hubDevice struct {
}

func NewModel() device.Devicer {
	return &hubDevice{}
}

func (h *hubDevice) GetConfig() device.Config {
	return device.Config{
		Model:   "hub",
		Parents: []string{"hub"},
		Flags:   device.FlagProgenitive | device.FlagHttpPortMust,
		State:   h,
		FS:      &fs,
		Targets: []string{"x86-64", "rpi", "koyeb"},
		BgColor: "sunflower",
		FgColor: "black",
	}
}

func (h *hubDevice) Setup() error                { return nil }
func (h *hubDevice) Poll(pkt *device.Packet)     {}
func (h *hubDevice) DemoSetup() error            { return h.Setup() }
func (h *hubDevice) DemoPoll(pkt *device.Packet) { h.Poll(pkt) }
