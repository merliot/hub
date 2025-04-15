package gps

import (
	"time"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/gps"
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
