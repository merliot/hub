//go:build !tinygo

package prostar

import (
	"embed"
	"text/template"

	"github.com/merliot/hub"
)

//go:embed images *.go template
var fs embed.FS

func (p *prostar) GetConfig() hub.Config {
	return hub.Config{
		Model:      "prostar",
		State:      p,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi", "nano-rp2040"},
		BgColor:    "sky",
		FgColor:    "black",
		PollPeriod: pollPeriod,
		PacketHandlers: hub.PacketHandlers{
			"/update-status":     &hub.PacketHandler[Status]{p.save},
			"/update-system":     &hub.PacketHandler[System]{p.save},
			"/update-controller": &hub.PacketHandler[Controller]{p.save},
			"/update-battery":    &hub.PacketHandler[Battery]{p.save},
			"/update-load":       &hub.PacketHandler[Load]{p.save},
			"/update-array":      &hub.PacketHandler[Array]{p.save},
			"/update-daily":      &hub.PacketHandler[Daily]{p.save},
		},
		FuncMap: template.FuncMap{
			"chargeState": p.chargeState,
			"loadState":   p.loadState,
		},
	}
}

func (p *prostar) chargeState(state uint16) string {
	switch state {
	case 0:
		return "START"
	case 1:
		return "NIGHT CHECK"
	case 2:
		return "DISCONNECT"
	case 3:
		return "NIGHT"
	case 4:
		return "FAULT"
	case 5:
		return "BULK"
	case 6:
		return "ABSORPTION"
	case 7:
		return "FLOAT"
	case 8:
		return "EQUALIZE"
	}
	return "??"
}

func (p *prostar) loadState(state uint16) string {
	switch state {
	case 0:
		return "START"
	case 1:
		return "LOAD ON"
	case 2:
		return "LVD WARNING"
	case 3:
		return "LVD"
	case 4:
		return "FAULT"
	case 5:
		return "DISCONNECT"
	case 6:
		return "LOAD OFF"
	case 7:
		return "OVERRIDE"
	}
	return "??"
}
