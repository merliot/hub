package gadget

import (
	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/led"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type gadget struct {
	Bottles   int // Bottles on the wall
	Restock   int // Restock countdown timer
	FullCount int `schema:"desc=Full bottle count"`
	io.Led
}

type msgTakeone struct {
}

func (m msgTakeone) Desc() string {
	return "Take one down, pass it around"
}

func NewModel() device.Devicer {
	return &gadget{FullCount: 99}
}

func (g *gadget) Setup() error {
	g.Bottles = g.FullCount
	g.Restock = 70
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
		pkt.SetPath("update").Marshal(g).BroadcastUp()
	}
}

func (g *gadget) takeone(pkt *device.Packet) {
	if g.Bottles > 0 {
		g.Bottles--
		pkt.SetPath("update").Marshal(g).BroadcastUp()
	}
}

func (g *gadget) update(pkt *device.Packet) {
	pkt.Unmarshal(g).BroadcastUp()
}

func (g *gadget) DemoSetup() error            { return g.Setup() }
func (g *gadget) DemoPoll(pkt *device.Packet) { g.Poll(pkt) }

func (g *gadget) Detail() Node {
	return Div(
		Class("flex flex-col"),
		ID("gadget-bottles"),
		Div(
			Class("flex flex-row w-full items-center justify-evenly"),
			Span(Class("text-6xl mx-4"), Textf("%d", g.Bottles)),
			Div(
				Class("flex flex-col items-center"),
				Span(Text("Bottles of Beer on the Wall")),
				If(g.Restock <= 60,
					Span(Class("text-sm"), Textf("[Restocking in %ds]", g.Restock)),
				),
			),
		),
		Div(
			Class("flex flex-row justify-end"),
			Button(
				Attr("hx-post", "/device/gadget/reboot"),
				Attr("hx-swap", "none"),
				Text("Reboot"),
			),
			Button(
				Attr("hx-post", "/device/gadget/takeone"),
				Attr("hx-swap", "none"),
				Text("Take One"),
			),
		),
	)
}

func (g *gadget) Overview() Node {
	return Div(
		Class("flex flex-col items-center"),
		Span(Class("text-lg"), Textf("%d Bottles", g.Bottles)),
	)
}
