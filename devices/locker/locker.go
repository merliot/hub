package locker

import (
	"github.com/merliot/hub/pkg/device"
)

type locker struct {
	Secret string `json:"-"`
}

func NewModel() device.Devicer {
	return &locker{}
}

func (l *locker) GetConfig() device.Config {
	return device.Config{
		Model:   "locker",
		State:   l,
		FS:      &fs,
		Targets: []string{"x86-64", "rpi", "nano-rp2040", "wioterminal", "pyportal"},
		BgColor: "mars",
		FgColor: "black",
	}
}

func (l *locker) Setup() error                { return nil }
func (l *locker) Poll(pkt *device.Packet)     {}
func (l *locker) DemoSetup() error            { return l.Setup() }
func (l *locker) DemoPoll(pkt *device.Packet) { l.Poll(pkt) }
