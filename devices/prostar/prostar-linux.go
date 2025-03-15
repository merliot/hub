//go:build !tinygo

package prostar

import (
	"embed"

	"github.com/merliot/hub/pkg/device"
)

//go:embed images *.go template
var fs embed.FS

func (p *prostar) GetConfig() device.Config {
	return device.Config{
		Model:      "prostar",
		Parents:    []string{"hub"},
		State:      p,
		FS:         &fs,
		Targets:    []string{"x86-64", "rpi", "nano-rp2040"},
		BgColor:    "sky",
		FgColor:    "black",
		PollPeriod: pollPeriod,
		PacketHandlers: device.PacketHandlers{
			"/update-status":     &device.PacketHandler[Status]{p.save},
			"/update-system":     &device.PacketHandler[System]{p.save},
			"/update-controller": &device.PacketHandler[Controller]{p.save},
			"/update-battery":    &device.PacketHandler[Battery]{p.save},
			"/update-load":       &device.PacketHandler[Load]{p.save},
			"/update-array":      &device.PacketHandler[Array]{p.save},
			"/update-daily":      &device.PacketHandler[Daily]{p.save},
		},
		FuncMap: device.FuncMap{
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
