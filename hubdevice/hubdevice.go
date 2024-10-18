package hubdevice

import (
	"embed"

	"github.com/merliot/hub"
)

//go:embed images *.go template
var fs embed.FS

type Hub struct {
}

func NewModel() hub.Devicer {
	return &Hub{}
}

func (h *Hub) GetConfig() hub.Config {
	return hub.Config{
		Model:   "hub",
		Flags:   hub.FlagProgenitive | hub.FlagWantsHttpPort,
		State:   h,
		FS:      &fs,
		Targets: []string{"x86-64", "rpi"},
		BgColor: "sunflower",
	}
}

func (h *Hub) GetHandlers() hub.Handlers {
	return hub.Handlers{}
}

func (h *Hub) Setup() error             { return nil }
func (h *Hub) Poll(pkt *hub.Packet)     {}
func (h *Hub) DemoSetup() error         { return nil }
func (h *Hub) DemoPoll(pkt *hub.Packet) {}
