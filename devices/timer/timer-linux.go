//go:build !tinygo

package timer

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/merliot/hub/pkg/device"
)

//go:embed *.go images template
var fs embed.FS

func (t *timer) GetConfig() device.Config {
	return device.Config{
		Model:      "timer",
		Parents:    []string{"hub"},
		State:      t,
		FS:         &fs,
		Targets:    []string{"rpi", "nano-rp2040"},
		PollPeriod: time.Second,
		BgColor:    "gray",
		FgColor:    "black",
		PacketHandlers: device.PacketHandlers{
			"/update": &device.PacketHandler[msgUpdate]{t.update},
		},
		FuncMap: template.FuncMap{
			"timeZones": timeZones,
			"update":    t.makeStatus,
		},
		APIs: device.APIs{
			"GET /update": t.status,
		},
	}
}

func (t *timer) makeStatus() string {
	if t.On {
		return "Turning Off in " + time.Until(t.stopTime).String()
	} else {
		return "Turning On in " + time.Since(t.startTime).String()
	}
}

func (t *timer) status(w http.ResponseWriter, r *http.Request) {
	status := t.makeStatus()
	resp := "<span>" + status + "</span>"
	w.Write([]byte(resp))
}
