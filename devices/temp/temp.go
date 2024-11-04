package temp

import (
	"fmt"
	"html/template"
	"time"

	"github.com/merliot/hub"
	io "github.com/merliot/hub/io/temp"
)

type temp struct {
	Temperature float32 // deg C
	Humidity    float32 // %
	Sensor      string
	Gpio        string
	TempUnits   string
	io.Temp
}

type msgUpdate struct {
	Temperature float32
	Humidity    float32
}

func NewModel() hub.Devicer {
	return &temp{}
}

func (t *temp) GetConfig() hub.Config {
	return hub.Config{
		Model:      "temp",
		State:      t,
		FS:         &fs,
		Targets:    []string{"rpi", "nano-rp2040", "wioterminal"},
		BgColor:    "orange",
		FgColor:    "black",
		PollPeriod: 5 * time.Second,
		PacketHandlers: hub.PacketHandlers{
			"/update": &hub.PacketHandler[msgUpdate]{t.update},
		},
		FuncMap: template.FuncMap{
			"tempf": t.tempf,
			"humf":  t.humf,
		},
	}
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

func (t *temp) update(pkt *hub.Packet) {
	pkt.Unmarshal(t).RouteUp()
}

func (t *temp) Setup() error {
	return t.Temp.Setup(t.Sensor, t.Gpio)
}

func (t *temp) Poll(pkt *hub.Packet) {
	temp, hum, err := t.Read()
	if err != nil {
		hub.LogError("Temp device poll read", "err", err)
		return
	}
	t.Temperature = temp
	t.Humidity = hum
	var msg = msgUpdate{temp, hum}
	pkt.SetPath("/update").Marshal(&msg).RouteUp()
}
