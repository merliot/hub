package gps

import (
	"fmt"
	"time"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/gps"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type gps struct {
	Lat        float64
	Long       float64
	Radius     float64 `schema:"desc=Radius for updates,units=meters"`
	PollPeriod uint    `schema:"desc=Poll period,units=seconds"`
	io.Gps
}

type updateMsg struct {
	Lat  float64
	Long float64
}

func NewModel() device.Devicer {
	return &gps{Radius: 50, PollPeriod: 30}
}

func (g *gps) GetConfig() device.Config {
	return device.Config{
		Model:      "gps",
		Parents:    []string{"hub"},
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi", "nano-rp2040"},
		BgColor:    "green",
		FgColor:    "black",
		PollPeriod: time.Second * time.Duration(g.PollPeriod),
		PacketHandlers: device.PacketHandlers{
			"update": &device.PacketHandler[updateMsg]{g.update},
		},
	}
}

func (g *gps) update(pkt *device.Packet) {
	pkt.Unmarshal(g).BroadcastUp()
}

func (g *gps) Poll(pkt *device.Packet) {
	lat, long, _ := g.Location()
	dist := io.Distance(lat, long, g.Lat, g.Long)
	if dist >= g.Radius {
		var up = updateMsg{lat, long}
		g.Lat, g.Long = lat, long
		pkt.SetPath("update").Marshal(&up).BroadcastUp()
	}
}

// Implement Detail() and Overview() for gps
func (g *gps) Detail() Node {
	return Div(
		Class("flex flex-col items-center"),
		Div(
			Class("w-full h-52 rounded-2xl overflow-hidden"),
			Attr("id", "gps-map"),
		),
		// Placeholder for map script
		Script(Raw(`/* Map script placeholder: center at Lat: `+f64(g.Lat)+`, Long: `+f64(g.Long)+` */`)),
	)
}

func (g *gps) Overview() Node {
	return Div(
		Class("flex flex-row items-center justify-evenly"),
		Attr("id", "gps-overview"),
		Span(Textf("Lat: %.5f°, Long: %.5f°", g.Lat, g.Long)),
	)
}

// Helper to format float64 as string
func f64(f float64) string {
	return fmt.Sprintf("%.5f", f)
}
