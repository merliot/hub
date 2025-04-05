//go:build !tinygo

package timer

import (
	"embed"
	"time"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var fs embed.FS

func (t *timer) GetConfig() device.Config {
	return device.Config{
		Model:      "timer",
		Parents:    []string{"hub"},
		State:      t,
		FS:         &fs,
		Targets:    []string{"rpi", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:    "butterscotch",
		FgColor:    "black",
		PacketHandlers: device.PacketHandlers{
			"/update": &device.PacketHandler[msgUpdate]{t.update},
		},
	}
}
