package locker

import (
	"github.com/merliot/hub/pkg/device"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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
		Parents: []string{"hub"},
		State:   l,
		FS:      &fs,
		Targets: []string{"x86-64", "rpi", "nano-rp2040", "pyportal"},
		BgColor: "mars",
		FgColor: "black",
	}
}

func (l *locker) Setup() error                { return nil }
func (l *locker) Poll(pkt *device.Packet)     {}
func (l *locker) DemoSetup() error            { return l.Setup() }
func (l *locker) DemoPoll(pkt *device.Packet) { l.Poll(pkt) }

// Implement Detail() and Overview() for locker
func (l *locker) Detail() Node {
	// We cannot check online status directly from locker struct, so we show both cases for now.
	// In real integration, parent device struct should pass online status to this method.
	return Div(
		Class("flex flex-col m-4"),
		Attr("id", "locker-secret"),
		H4(Text("SECRET")),
		// This is a placeholder: in real use, online status should be passed in
		Div(
			Class("flex flex-row justify-center"),
			P(Text("[ OFFLINE -- SECRET HIDDEN ]")),
		),
	)
}

func (l *locker) Overview() Node {
	// No overview content for locker
	return Group(nil)
}
