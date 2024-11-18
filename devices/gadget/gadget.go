package gadget

import (
	"fmt"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/led"
)

type gadget struct {
	Bottles   int // Bottles on the wall
	Restock   int // Restock countdown timer
	fullCount int // Full bottle count
	io.Led
}

func NewModel() device.Devicer {
	return &gadget{Bottles: 99, Restock: 70}
}

func (g *gadget) Setup() error {
	if err := g.Led.Setup(); err != nil {
		return err
	}
	if g.Bottles < 1 {
		return fmt.Errorf("gadget Bottles < 1")
	}
	g.fullCount = g.Bottles
	return nil
}

func (g *gadget) Poll(pkt *device.Packet) {
	if g.Bottles < g.fullCount {
		if g.Restock == 1 {
			g.Bottles = g.fullCount
			g.Restock = 70
			g.Led.Off()
		} else {
			g.Restock--
			g.Led.On()
		}
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *gadget) takeone(pkt *device.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		pkt.SetPath("/update").Marshal(g).RouteUp()
	}
}

func (g *gadget) update(pkt *device.Packet) {
	pkt.Unmarshal(g).RouteUp()
}

func (g *gadget) DemoSetup() error            { return g.Setup() }
func (g *gadget) DemoPoll(pkt *device.Packet) { g.Poll(pkt) }
