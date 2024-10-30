package relays

import (
	"fmt"
	"net/url"

	"github.com/merliot/hub"
	io "github.com/merliot/hub/io/relay"
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
	Relay int
}

type msgClicked struct {
	Relay int
	State bool
}

func NewModel() hub.Devicer {
	return &relays{}
}

func (r *relays) GetConfig() hub.Config {
	return hub.Config{
		Model:   "relays",
		State:   r,
		FS:      &fs,
		Targets: []string{"rpi", "nano-rp2040", "wioterminal"},
		BgColor: "ice",
		FgColor: "black",
	}
}

func (r *relays) GetHandlers() hub.Handlers {
	return hub.Handlers{
		"/click":   &hub.Handler[msgClick]{r.click},
		"/clicked": &hub.Handler[msgClicked]{r.clicked},
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

func (r *relays) Poll(pkt *hub.Packet) {
}

func (r *relays) click(pkt *hub.Packet) {
	var click msgClick
	pkt.Unmarshal(&click)
	relay := &r.Relays[click.Relay]
	relay.Set(!relay.State)
	var clicked = msgClicked{click.Relay, relay.State}
	pkt.SetPath("/clicked").Marshal(&clicked).RouteUp()
}

func (r *relays) clicked(pkt *hub.Packet) {
	var clicked msgClicked
	pkt.Unmarshal(&clicked)
	relay := &r.Relays[clicked.Relay]
	relay.Set(clicked.State)
	pkt.RouteUp()
}

func (r *relays) DemoSetup() error         { return r.Setup() }
func (r *relays) DemoPoll(pkt *hub.Packet) { r.Poll(pkt) }
