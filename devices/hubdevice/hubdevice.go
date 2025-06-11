package hubdevice

import (
	"embed"

	"github.com/merliot/hub/pkg/device"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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
		Model:       "hub",
		Parents:     []string{"hub"},
		Flags:       device.FlagProgenitive | device.FlagHttpPortMust,
		State:       h,
		FS:          &fs,
		Targets:     []string{"x86-64", "rpi", "koyeb"},
		BgColor:     "bg-yellow-300",
		FgColor:     "text-black",
		BorderColor: "border-yellow-300",
		InitialView: "detail",
	}
}

func (h *hubDevice) Setup() error                { return nil }
func (h *hubDevice) Poll(pkt *device.Packet)     {}
func (h *hubDevice) DemoSetup() error            { return h.Setup() }
func (h *hubDevice) DemoPoll(pkt *device.Packet) { h.Poll(pkt) }

func (h *hubDevice) Detail() Node {
	// This is a placeholder. In real use, you would check if this device is root and has children.
	return Div(
		Class("flex flex-row justify-end"),
		Span(Class("m-4"), Text("No devices. Click New to add devices.")),
		// TODO: Add a real New button component here
	)
}

func (h *hubDevice) Overview() Node {
	return Div(
		Text("Hub Device Overview (placeholder)"),
	)
}
