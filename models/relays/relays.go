package relays

import (
	"embed"
	"net/http"
	"os"
	"os/signal"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js index.html
var fs embed.FS

type Relays struct {
	*common.Common
	relays [4]*gpio.RelayDriver
	States [4]bool
}

type MsgClick struct {
	Path  string
	Relay int
	State bool
}

func New(id, model, name string) dean.Thinger {
	println("NEW RELAYS")
	return &Relays{
		Common: common.New(id, model, name).(*common.Common),
	}
}

func (r *Relays) save(msg *dean.Msg) {
	msg.Unmarshal(r)
}

func (r *Relays) getState(msg *dean.Msg) {
	r.Path = "state"
	msg.Marshal(r).Reply()
}

func (r *Relays) click(msg *dean.Msg) {
	var msgClick MsgClick
	msg.Unmarshal(&msgClick)
	r.States[msgClick.Relay] = msgClick.State
	if r.IsMetal() {
		if msgClick.State {
			r.relays[msgClick.Relay].On()
		} else {
			r.relays[msgClick.Relay].Off()
		}
	}
	msg.Broadcast()
}

func (r *Relays) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     r.save,
		"get/state": r.getState,
		"click":     r.click,
	}
}

func (r *Relays) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.API(fs, w, req)
}

func (r *Relays) Run(i *dean.Injector) {

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	r.relays[0] = gpio.NewRelayDriver(adaptor, "31") // GPIO 6
	r.relays[1] = gpio.NewRelayDriver(adaptor, "33") // GPIO 13
	r.relays[2] = gpio.NewRelayDriver(adaptor, "35") // GPIO 19
	r.relays[3] = gpio.NewRelayDriver(adaptor, "37") // GPIO 26

	for _, relay := range r.relays {
		relay.Start()
		relay.Off()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		for _, relay := range r.relays {
			relay.Off()
		}
	}
}
