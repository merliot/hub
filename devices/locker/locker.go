package locker

import (
	"github.com/merliot/hub"
)

type locker struct {
	Secret string `json:"-"`
}

func NewModel() hub.Devicer {
	return &locker{}
}

func (l *locker) GetConfig() hub.Config {
	return hub.Config{
		Model:   "locker",
		State:   l,
		FS:      &fs,
		Targets: []string{"x86-64", "rpi", "nano-rp2040", "wioterminal"},
		BgColor: "mars",
		FgColor: "black",
	}
}

func (l *locker) GetHandlers() hub.Handlers { return hub.Handlers{} }
func (l *locker) Setup() error              { return nil }
func (l *locker) Poll(pkt *hub.Packet)      {}
func (l *locker) DemoSetup() error          { return l.Setup() }
func (l *locker) DemoPoll(pkt *hub.Packet)  { l.Poll(pkt) }
