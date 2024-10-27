package hubdevice

import (
	"embed"

	"github.com/merliot/hub"
)

//go:embed images *.go template
var fs embed.FS

type hubDevice struct {
}

func NewModel() hub.Devicer {
	return &hubDevice{}
}

func (h *hubDevice) GetConfig() hub.Config {
	return hub.Config{
		Model:   "hub",
		Flags:   hub.FlagProgenitive | hub.FlagWantsHttpPort,
		State:   h,
		FS:      &fs,
		Targets: []string{"x86-64", "rpi"},
		BgColor: "sunflower",
		FgColor: "black",
	}
}

func (h *hubDevice) GetHandlers() hub.Handlers {
	return hub.Handlers{}
}

func (h *hubDevice) Setup() error             { return nil }
func (h *hubDevice) Poll(pkt *hub.Packet)     {}
func (h *hubDevice) DemoSetup() error         { return nil }
func (h *hubDevice) DemoPoll(pkt *hub.Packet) {}
