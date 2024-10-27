package relays

import (
	"github.com/merliot/hub"
	io "github.com/merliot/hub/io/relay"
)

type relays struct {
	Relays [4]io.Relay
}

type MsgClick struct {
	Relay int
}

type MsgClicked struct {
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
		"/click":   &hub.Handler[MsgClick]{r.click},
		"/clicked": &hub.Handler[MsgClicked]{r.clicked},
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
	var click MsgClick
	pkt.Unmarshal(&click)
	relay := &r.Relays[click.Relay]
	relay.Set(!relay.State)
	var clicked = MsgClicked{click.Relay, relay.State}
	pkt.SetPath("/clicked").Marshal(&clicked).RouteUp()
}

func (r *relays) clicked(pkt *hub.Packet) {
	var clicked MsgClicked
	pkt.Unmarshal(&clicked)
	relay := &r.Relays[clicked.Relay]
	relay.Set(clicked.State)
	pkt.RouteUp()
}

func (r *relays) DemoSetup() error         { return r.Setup() }
func (r *relays) DemoPoll(pkt *hub.Packet) { r.Poll(pkt) }
