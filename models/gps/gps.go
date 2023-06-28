package gps

import (
	"embed"
	"net/http"
	"os"
	"os/signal"

	"github.com/merliot/dean"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js index.html
var fs embed.FS

type Gps struct {
	dean.Thing
	dean.ThingMsg
	Identity Identity
	relays [4]*gpio.RelayDriver
	States [4]bool
}

type Identity struct {
	Id    string
	Model string
	Name  string
}

type MsgClick struct {
	Path  string
	Relay int
	State bool
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	return &Gps{
		Thing: dean.NewThing(id, model, name),
		Identity: Identity{id, model, name},
	}
}

func (g *Gps) saveState(msg *dean.Msg) {
	msg.Unmarshal(g)
}

func (g *Gps) getState(msg *dean.Msg) {
	g.Path = "state"
	msg.Marshal(g).Reply()
}

func (g *Gps) click(msg *dean.Msg) {
	var msgClick MsgClick
	msg.Unmarshal(&msgClick)
	g.States[msgClick.Relay] = msgClick.State
	if g.IsMetal() {
		if msgClick.State {
			g.relays[msgClick.Relay].On()
		} else {
			g.relays[msgClick.Relay].Off()
		}
	}
	msg.Broadcast()
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.saveState,
		"get/state": g.getState,
		"click":     g.click,
	}
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.ServeFS(fs, w, r)
}

func (g *Gps) Run(i *dean.Injector) {

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	g.relays[0] = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	g.relays[1] = gpio.NewRelayDriver(adaptor, "33") // GPIO 13
	g.relays[2] = gpio.NewRelayDriver(adaptor, "35") // GPIO 19
	g.relays[3] = gpio.NewRelayDriver(adaptor, "37") // GPIO 26

	for _, relay := range g.relays {
		relay.Start()
		relay.Off()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		for _, relay := range g.relays {
			relay.Off()
		}
	}
}
