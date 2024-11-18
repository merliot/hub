//go:build tinygo

package gadget

import (
	"time"

	"github.com/merliot/hub/pkg/device"
)

func (g *gadget) GetConfig() device.Config {
	return device.Config{
		Model:      "gadget",
		State:      g,
		PollPeriod: time.Second,
		PacketHandlers: device.PacketHandlers{
			"/takeone": &device.PacketHandler[device.NoMsg]{g.takeone},
			"/update":  &device.PacketHandler[gadget]{g.update},
		},
	}
}
