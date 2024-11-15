//go:build !tinygo

package temp

import (
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/merliot/hub"
)

//go:embed *.go images template
var fs embed.FS

func (t *temp) GetConfig() hub.Config {
	return hub.Config{
		Model:      "temp",
		State:      t,
		FS:         &fs,
		Targets:    []string{"nano-rp2040"},
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

func (t *temp) points(series, originX, originY, width, height uint, minY, maxY int) string {
	var points []string

	stepX := float32(width) / float32(historyRecs-1)
	scaleY := float32(height) / float32(maxY-minY)

	for i, rec := range t.History {
		pos := len(t.History) - 1 - i
		x := float32(originX) + float32(width) - (float32(pos) * stepX)
		y := float32(originY) + float32(height) - (rec[series] * scaleY)
		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
	}

	// points="0,120 20,60 40,80 60,20"
	return strings.Join(points, " ")
}

func (t *temp) tempf() string {
	return fmt.Sprintf("%.1f", t.Temperature)
}

func (t *temp) humf() string {
	return fmt.Sprintf("%.1f", t.Humidity)
}
