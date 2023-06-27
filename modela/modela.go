package modela

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js images index.html
var fs embed.FS

type Modela struct {
	dean.Thing
	dean.ThingMsg
	relays [4]*gpio.RelayDriver
	States [4]bool
}

type MsgClick struct {
	Path  string
	Relay int
	State bool
}

func New(id, model, name string) dean.Thinger {
	println("NEW MODELA")
	return &Modela{
		Thing: dean.NewThing(id, model, name),
	}
}

func (m *Modela) saveState(msg *dean.Msg) {
	msg.Unmarshal(m)
}

func (m *Modela) getState(msg *dean.Msg) {
	m.Path = "state"
	msg.Marshal(m).Reply()
}

func (m *Modela) click(msg *dean.Msg) {
	var msgClick MsgClick
	msg.Unmarshal(&msgClick)
	m.States[msgClick.Relay] = msgClick.State
	println(m.IsMetal(), msgClick.State)
	if m.IsMetal() {
		if msgClick.State {
			m.relays[msgClick.Relay].On()
		} else {
			m.relays[msgClick.Relay].Off()
		}
	}
	msg.Broadcast()
}

func (m *Modela) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     m.saveState,
		"get/state": m.getState,
		"click":     m.click,
	}
}

func (m *Modela) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeFS(fs, w, r)
}

func (m *Modela) Run(i *dean.Injector) {

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	m.relays[0] = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	m.relays[1] = gpio.NewRelayDriver(adaptor, "33") // GPIO 13
	m.relays[2] = gpio.NewRelayDriver(adaptor, "35") // GPIO 19
	m.relays[3] = gpio.NewRelayDriver(adaptor, "37") // GPIO 26

	for _, relay := range m.relays {
		relay.Start()
		relay.On()
	}

	select {}
}
