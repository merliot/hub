package temp

import (
	"time"

	"github.com/merliot/hub"
	io "github.com/merliot/hub/io/temp"
)

var (
	pollPeriod = 10 * time.Second
)

type Record []float32

type History []Record

type temp struct {
	Temperature float32 // deg C
	Humidity    float32 // %
	Sensor      string
	Gpio        string
	TempUnits   string
	History
	io.Temp
}

type msgUpdate struct {
	Temperature float32
	Humidity    float32
}

func NewModel() hub.Devicer {
	return &temp{
		History: []Record{},
	}
}

func (t *temp) addRecord() {
	if len(t.History) >= 60 {
		// Remove the oldest
		t.History = t.History[1:]
	}
	// Add the new
	rec := Record{t.Temperature, t.Humidity}
	t.History = append(t.History, rec)
}

func (t *temp) update(pkt *hub.Packet) {
	t.addRecord()
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
	t.addRecord()
	var msg = msgUpdate{temp, hum}
	pkt.SetPath("/update").Marshal(&msg).RouteUp()
}
