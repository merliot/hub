package buttons

import (
	"fmt"
	"net/url"
	"time"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/button"
)

type buttons struct {
	Buttons [4]io.Button `schema:"desc=Buttons"`
}

func (b *buttons) Decode(values url.Values) error {
	// Custom decoder since tinygo's reflect.ArrayOf() panics
	if len(values) != 0 {
		for i := range b.Buttons {
			button := &b.Buttons[i]
			nameKey := fmt.Sprintf("Buttons[%d].Name", i)
			button.Name = values.Get(nameKey)
			gpioKey := fmt.Sprintf("Buttons[%d].Gpio", i)
			button.Gpio = values.Get(gpioKey)
		}
	}
	return nil
}

type msgUpdate struct {
	Button int
	State  bool
}

func NewModel() device.Devicer {
	return &buttons{}
}

func (b *buttons) GetConfig() device.Config {
	return device.Config{
		Model:      "buttons",
		Parents:    []string{"hub"},
		State:      b,
		FS:         &fs,
		Targets:    []string{"rpi", "nano-rp2040"},
		PollPeriod: 10 * time.Millisecond,
		BgColor:    "gold",
		FgColor:    "black",
		PacketHandlers: device.PacketHandlers{
			"update": &device.PacketHandler[msgUpdate]{b.update},
		},
	}
}

func (b *buttons) Setup() error {
	for i := range b.Buttons {
		button := &b.Buttons[i]
		if err := button.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (b *buttons) Poll(pkt *device.Packet) {
	for i := range b.Buttons {
		button := &b.Buttons[i]
		last := button.State
		curr := button.Get()
		if curr != last {
			var update = msgUpdate{i, curr}
			pkt.SetPath("update").Marshal(&update).BroadcastUp()
		}
	}
}

func (b *buttons) update(pkt *device.Packet) {
	var update msgUpdate
	pkt.Unmarshal(&update)
	button := &b.Buttons[update.Button]
	button.State = update.State
	pkt.BroadcastUp()
}
