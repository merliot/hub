package gps

import (
	"time"

	"github.com/merliot/hub"
	io "github.com/merliot/hub/io/gps"
)

type gps struct {
	Lat        float64
	Long       float64
	Radius     float64 // units: meters
	PollPeriod uint    // units: seconds
	io.Gps
}

type updateMsg struct {
	Lat  float64
	Long float64
}

func NewModel() hub.Devicer {
	return &gps{Radius: 50, PollPeriod: 30}
}

func (g *gps) GetConfig() hub.Config {
	return hub.Config{
		Model:      "gps",
		State:      g,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi", "nano-rp2040", "wioterminal"},
		BgColor:    "green",
		FgColor:    "black",
		PollPeriod: time.Second * time.Duration(g.PollPeriod),
	}
}

func (g *gps) GetHandlers() hub.Handlers {
	return hub.Handlers{
		"/update": &hub.Handler[updateMsg]{g.update},
	}
}

func (g *gps) Poll(pkt *hub.Packet) {
	lat, long, _ := g.Location()
	dist := io.Distance(lat, long, g.Lat, g.Long)
	if dist >= g.Radius {
		var up = updateMsg{lat, long}
		g.Lat, g.Long = lat, long
		pkt.SetPath("/update").Marshal(&up).RouteUp()
	}
}

func (g *gps) update(pkt *hub.Packet) {
	pkt.Unmarshal(g).RouteUp()
}
