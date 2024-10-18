package relays

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/io/relay"
)

type Relays struct {
	Relays [4]relay.Relay
}

type MsgClick struct {
	Relay int
}

type MsgClicked struct {
	Relay int
	State bool
}

func NewModel() hub.Devicer {
	return &Relays{}
}

func (r *Relays) GetConfig() hub.Config {
	return hub.Config{
		Model:   "relays",
		State:   r,
		FS:      &fs,
		Targets: []string{"rpi", "nano-rp2040", "wioterminal"},
		BgColor: "orange",
	}
}

func (r *Relays) GetHandlers() hub.Handlers {
	return hub.Handlers{
		"/state":   &hub.Handler[Relays]{r.state},
		"/click":   &hub.Handler[MsgClick]{r.click},
		"/clicked": &hub.Handler[MsgClicked]{r.clicked},
	}
}

func (r *Relays) Setup() error {
	for i := range r.Relays {
		relay := &r.Relays[i]
		if err := relay.Setup(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Relays) Poll(pkt *hub.Packet) {
}

func (r *Relays) state(pkt *hub.Packet) {
	pkt.Unmarshal(r).RouteUp()
}

func (r *Relays) click(pkt *hub.Packet) {
	var click MsgClick
	pkt.Unmarshal(&click)
	relay := &r.Relays[click.Relay]
	relay.Set(!relay.State)
	var clicked = MsgClicked{click.Relay, relay.State}
	pkt.SetPath("/clicked").Marshal(&clicked).RouteUp()
}

func (r *Relays) clicked(pkt *hub.Packet) {
	var clicked MsgClicked
	pkt.Unmarshal(&clicked)
	relay := &r.Relays[clicked.Relay]
	relay.Set(clicked.State)
	pkt.RouteUp()
}

func (r *Relays) DemoSetup() error            { return r.Setup() }
func (r *Relays) DemoPoll(pkt *hub.Packet) { r.Poll(pkt) }
