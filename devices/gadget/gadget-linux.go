//go:build !tinygo

package gadget

import (
	"embed"
	"time"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var fs embed.FS

func (g *gadget) GetConfig() device.Config {
	return device.Config{
		Model:      "gadget",
		Parents:    []string{"hub"},
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:    "african-violet",
		FgColor:    "black",
		PacketHandlers: device.PacketHandlers{
			"/takeone": &device.PacketHandler[msgTakeone]{g.takeone},
			"update":   &device.PacketHandler[gadget]{g.update},
		},
	}
}
