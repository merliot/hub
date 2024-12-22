package moistbase

import (
	"embed"
	"time"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go template
var fs embed.FS

type moistbase struct {
}

func NewModel() device.Devicer {
	return &moistbase{}
}

func (m *moistbase) GetConfig() device.Config {
	return device.Config{
		Model:      "moistbase",
		Parents:    []string{"hub"},
		State:      m,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi"},
		PollPeriod: time.Second,
		BgColor:    "moonlit-violet",
		FgColor:    "black",
	}
}

func (m *moistbase) Setup() error                { return nil }
func (m *moistbase) Poll(pkt *device.Packet)     {}
func (m *moistbase) DemoSetup() error            { return m.Setup() }
func (m *moistbase) DemoPoll(pkt *device.Packet) { m.Poll(pkt) }
