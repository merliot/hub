//go:build tinygo

package gadget

import (
	"time"

	"github.com/merliot/hub"
)

func (g *gadget) GetConfig() hub.Config {
	return hub.Config{
		Model:      "gadget",
		State:      g,
		PollPeriod: time.Second,
		PacketHandlers: hub.PacketHandlers{
			"/takeone": &hub.PacketHandler[hub.NoMsg]{g.takeone},
			"/update":  &hub.PacketHandler[gadget]{g.update},
		},
	}
}
