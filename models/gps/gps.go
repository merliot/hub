package gps

import (
	"embed"
	"net/http"
	"os"
	"os/signal"

	"github.com/merliot/dean"
	"gobot.io/x/gobot/platforms/raspi"
)

//go:embed css js index.html
var fs embed.FS

type Gps struct {
	dean.Thing
	dean.ThingMsg
	Identity Identity
}

type Identity struct {
	Id    string
	Model string
	Name  string
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

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.saveState,
		"get/state": g.getState,
	}
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.ServeFS(fs, w, r)
}

func (g *Gps) Run(i *dean.Injector) {

	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
	}
}
