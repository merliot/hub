package sense

import (
	"embed"
	"net/http"

	"github.com/merliot/dean"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js index.html
var fs embed.FS

type Sense struct {
	dean.Thing
	dean.ThingMsg
	Identity Identity
	sense [4]*gpio.RelayDriver
	States [4]bool
}

type Identity struct {
	Id    string
	Model string
	Name  string
}

func New(id, model, name string) dean.Thinger {
	println("NEW RELAYS")
	return &Sense{
		Thing: dean.NewThing(id, model, name),
		Identity: Identity{id, model, name},
	}
}

func (s *Sense) saveState(msg *dean.Msg) {
	msg.Unmarshal(s)
}

func (s *Sense) getState(msg *dean.Msg) {
	s.Path = "state"
	msg.Marshal(s).Reply()
}

func (s *Sense) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     s.saveState,
		"get/state": s.getState,
	}
}

func (s *Sense) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.ServeFS(fs, w, r)
}

func (s *Sense) Run(i *dean.Injector) {

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	select {}
}
