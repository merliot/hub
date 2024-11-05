//go:build !tinygo

package temp

import (
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/merliot/hub"
)

//go:embed *.go template
var fs embed.FS

func (t *temp) GetConfig() hub.Config {
	return hub.Config{
		Model:      "temp",
		State:      t,
		FS:         &fs,
		Targets:    []string{"nano-rp2040", "wioterminal"},
		BgColor:    "orange",
		FgColor:    "black",
		PollPeriod: pollPeriod,
		PacketHandlers: hub.PacketHandlers{
			"/update": &hub.PacketHandler[msgUpdate]{t.update},
		},
		FuncMap: template.FuncMap{
			"tempf":  t.tempf,
			"humf":   t.humf,
			"points": t.points,
		},
	}
}

func (t *temp) points(set string, originX, originY, width, height uint) string {
	var points []string
	for _, rec := range t.History {
		points = append(points, fmt.Sprintf("%.1f,%.1f", rec[0], rec[1]))
	}
	// points="0,120 20,60 40,80 60,20"
	return strings.Join(points, " ")
}

func (t *temp) tempf() string {
	value := t.Temperature
	if t.TempUnits == "F" {
		value = (value * 9.0 / 5.0) + 32.0
	}
	return fmt.Sprintf("%.1f", value)
}

func (t *temp) humf() string {
	return fmt.Sprintf("%.1f", t.Humidity)
}
