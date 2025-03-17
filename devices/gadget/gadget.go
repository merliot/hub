package gadget

import (
	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/led"
)

type gadget struct {
	Bottles   int // Bottles on the wall
	Restock   int // Restock countdown timer
	FullCount int // Full bottle count
	io.Led
}

func NewModel() device.Devicer {
	return &gadget{Restock: 70}
}

func (g *gadget) Setup() error {
	g.Bottles = g.FullCount
	return g.Led.Setup()
}

func (g *gadget) Poll(pkt *device.Packet) {
	if g.Bottles < g.FullCount {
		if g.Restock == 1 {
			g.Bottles = g.FullCount
			g.Restock = 70
			g.Led.Off()
		} else {
			g.Restock--
			g.Led.On()
		}
		pkt.SetPath("/update").Marshal(g).BroadcastUp()
	}
}

func (g *gadget) takeone(pkt *device.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		pkt.SetPath("/update").Marshal(g).BroadcastUp()
	}
}

func (g *gadget) update(pkt *device.Packet) {
	pkt.Unmarshal(g).BroadcastUp()
}

func (g *gadget) DemoSetup() error            { return g.Setup() }
func (g *gadget) DemoPoll(pkt *device.Packet) { g.Poll(pkt) }
