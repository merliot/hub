package gadget

import (
	"embed"
	"time"

	"github.com/merliot/hub"
)

//go:embed *.go template
var fs embed.FS

type Gadget struct {
	Bottles int
	Restock int
}

func NewModel() hub.Devicer {
	return &Gadget{Bottles: 99, Restock: 70}
}

func (g *Gadget) GetConfig() hub.Config {
	return hub.Config{
		Model:      "gadget",
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:    "african-violet",
	}
}

func (g *Gadget) GetHandlers() hub.Handlers {
	return hub.Handlers{
		"/state":   &hub.Handler[Gadget]{g.state},
		"/takeone": &hub.Handler[hub.NoMsg]{g.takeone},
		"/update":  &hub.Handler[Gadget]{g.state},
	}
}

func (g *Gadget) Setup() error { return nil }

func (g *Gadget) Poll(pkt *hub.Packet) {
	if g.Bottles < 99 {
		if g.Restock == 1 {
			g.Bottles = 99
			g.Restock = 70
		} else {
			g.Restock--
		}
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *Gadget) state(pkt *hub.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *Gadget) takeone(pkt *hub.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *Gadget) DemoSetup() error            { return g.Setup() }
func (g *Gadget) DemoPoll(pkt *hub.Packet) { g.Poll(pkt) }
