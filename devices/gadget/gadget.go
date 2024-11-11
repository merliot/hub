package gadget

import (
	"embed"
	"fmt"
	"time"

	"github.com/merliot/hub"
)

//go:embed *.go images template
var fs embed.FS

type gadget struct {
	Bottles   int // Bottles on the wall
	Restock   int // Restock countdown timer
	fullCount int // Full bottle count
}

func NewModel() hub.Devicer {
	return &gadget{Bottles: 99, Restock: 70}
}

func (g *gadget) GetConfig() hub.Config {
	return hub.Config{
		Model:      "gadget",
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:    "african-violet",
		FgColor:    "black",
		PacketHandlers: hub.PacketHandlers{
			"/takeone": &hub.PacketHandler[hub.NoMsg]{g.takeone},
			"/update":  &hub.PacketHandler[gadget]{g.update},
		},
	}
}

func (g *gadget) Setup() error {
	if g.Bottles < 1 {
		return fmt.Errorf("gadget Bottles < 1")
	}
	g.fullCount = g.Bottles
	return nil
}

func (g *gadget) Poll(pkt *hub.Packet) {
	if g.Bottles < g.fullCount {
		if g.Restock == 1 {
			g.Bottles = g.fullCount
			g.Restock = 70
		} else {
			g.Restock--
		}
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *gadget) takeone(pkt *hub.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *gadget) update(pkt *hub.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *gadget) DemoSetup() error         { return g.Setup() }
func (g *gadget) DemoPoll(pkt *hub.Packet) { g.Poll(pkt) }
