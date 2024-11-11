//go:build !tinygo

package gadget

import (
	"embed"
	"time"

	"github.com/merliot/hub"
)

//go:embed *.go images template
var fs embed.FS

func (g *gadget) GetConfig() hub.Config {
	return hub.Config{
		Model:      "gadget",
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:    "african-violet",
		FgColor:    "black",
		PacketHandlers: hub.PacketHandlers{
			"/takeone": &hub.PacketHandler[hub.NoMsg]{g.takeone},
			"/update":  &hub.PacketHandler[gadget]{g.update},
		},
	}
}
