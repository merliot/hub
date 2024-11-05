//go:build !tinygo

package prostar

import (
	"embed"

	"github.com/merliot/hub"
)

//go:embed *.go template
var fs embed.FS

func (p *prostar) GetConfig() hub.Config {
	return hub.Config{
		Model:      "prostar",
		State:      p,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi", "nano-rp2040"},
		BgColor:    "sky",
		FgColor:    "black",
		PollPeriod: pollPeriod,
		PacketHandlers: hub.PacketHandlers{
			"/update/status":     &hub.PacketHandler[Status]{p.save},
			"/update/system":     &hub.PacketHandler[System]{p.save},
			"/update/controller": &hub.PacketHandler[Controller]{p.save},
			"/update/battery":    &hub.PacketHandler[Battery]{p.save},
			"/update/load":       &hub.PacketHandler[Load]{p.save},
			"/update/array":      &hub.PacketHandler[Array]{p.save},
			"/update/daily":      &hub.PacketHandler[Daily]{p.save},
		},
	}
}
