package relays

import (
	"fmt"
	"net/url"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/relay"
)

type relays struct {
	Relays [4]io.Relay
}

func (r *relays) Decode(values url.Values) error {

	// We shouldn't need a custom URL values decoder here, but tinygo's
	// reflect.ArrayOf() panics when trying to decode URL values with an
	// array using the form decorder.
	//
	// TODO this func can go away if reflect.ArrayOf() is implemented in
	// tinygo.

	if len(values) != 0 {
		for i := range r.Relays {
			relay := &r.Relays[i]
			nameKey := fmt.Sprintf("Relays[%d].Name", i)
			relay.Name = values.Get(nameKey)
			gpioKey := fmt.Sprintf("Relays[%d].Gpio", i)
			relay.Gpio = values.Get(gpioKey)
		}
	}
	return nil
}

type msgClick struct {
	Relay int `mcp:"required,desc=Relay index"`
}

func (m msgClick) desc() string {
	return "Click (toggle) the relay"
}

type msgClicked struct {
	Relay int
	State bool
}

func NewModel() device.Devicer {
	return &relays{}
}

func (r *relays) GetConfig() device.Config {
	return device.Config{
		Model:   "relays",
		Parents: []string{"hub"},
		State:   r,
		FS:      &fs,
		Targets: []string{"rpi", "nano-rp2040"},
		BgColor: "ice",
		FgColor: "black",
		PacketHandlers: device.PacketHandlers{
			"/click":  &device.PacketHandler[msgClick]{r.click},
			"clicked": &device.PacketHandler[msgClicked]{r.clicked},
		},
	}
}

func (r *relays) Setup() error {
	for i := range r.Relays {
		relay := &r.Relays[i]
		if err := relay.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (r *relays) click(pkt *device.Packet) {
	var click msgClick
	pkt.Unmarshal(&click)
	relay := &r.Relays[click.Relay]
	relay.Set(!relay.State)
	var clicked = msgClicked{click.Relay, relay.State}
	pkt.SetPath("clicked").Marshal(&clicked).BroadcastUp()
}

func (r *relays) clicked(pkt *device.Packet) {
	var clicked msgClicked
	pkt.Unmarshal(&clicked)
	relay := &r.Relays[clicked.Relay]
	relay.Set(clicked.State)
	pkt.BroadcastUp()
}

func (r *relays) Poll(pkt *device.Packet)     {}
func (r *relays) DemoSetup() error            { return r.Setup() }
func (r *relays) DemoPoll(pkt *device.Packet) { r.Poll(pkt) }
