//go:build !tinygo

package timer

import (
	"embed"
	"html/template"
	"net/url"
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
			"timezones": func() []string { return timezones },
		},
	}
}

func toUTC(hhmm, timezone string) (string, error) {

	// Load the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	now := time.Now()
	year, month, day := now.Year(), now.Month(), now.Day()

	// Parse the HH:MM string in the given timezone
	t, err := time.ParseInLocation("15:04", hhmm, loc)
	if err != nil {
		return "", err
	}

	// Set the time to the fixed date in the timezone
	localTime := time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, loc)

	// Convert to UTC and format as HH:MM
	utcTime := localTime.UTC()
	return utcTime.Format("15:04"), nil
}

func (t *timer) Decode(values url.Values) (err error) {

	t.StartHHMM = values.Get("StartHHMM")
	t.StopHHMM = values.Get("StopHHMM")
	t.TZ = values.Get("TZ")
	t.Gpio = values.Get("Gpio")
	t.StartUTC, _ = toUTC(t.StartHHMM, t.TZ)
	t.StopUTC, _ = toUTC(t.StopHHMM, t.TZ)

	return
}
