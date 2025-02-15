package moist

import (
	"time"

	"github.com/merliot/hub/pkg/device"
)

type moist struct {
}

func NewModel() device.Devicer {
	return &moist{}
}

func (m *moist) GetConfig() device.Config {
	return device.Config{
		Model:      "moist",
		Parents:    []string{"moistbase"},
		State:      m,
		Targets:    []string{"x86-64", "rpi"},
		PollPeriod: time.Second,
		BgColor:    "moonlit-violet",
		FgColor:    "black",
	}
}

func (m *moist) Setup() error                { return nil }
func (m *moist) Poll(pkt *device.Packet)     {}
func (m *moist) DemoSetup() error            { return m.Setup() }
func (m *moist) DemoPoll(pkt *device.Packet) { m.Poll(pkt) }
