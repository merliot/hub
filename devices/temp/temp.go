package temp

import (
	"time"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/temp"
)

var (
	pollPeriod  = time.Hour
	historyRecs = 24
)

type Record []float32

type History []Record

type temp struct {
	Temperature float32 // deg F or C, depends on TempUnits
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

func NewModel() device.Devicer {
	return &temp{
		History: []Record{},
	}
}

func (t *temp) addRecord() {
	if len(t.History) >= historyRecs {
		// Remove the oldest
		t.History = t.History[1:]
	}
	// Add the new
	rec := Record{t.Temperature, t.Humidity}
	t.History = append(t.History, rec)
}

func (t *temp) update(pkt *device.Packet) {
	t.addRecord()
	pkt.Unmarshal(t).BroadcastUp()
}

func (t *temp) Setup() error {
	return t.Temp.Setup(t.Sensor, t.Gpio)
}

func (t *temp) Poll(pkt *device.Packet) {
	temp, hum, err := t.Read()
	if err != nil {
		device.LogError("Temp device poll read", "err", err)
		return
	}
	if t.TempUnits == "F" {
		// Convert from Celsius
		temp = (temp * 9.0 / 5.0) + 32.0
	}
	t.Temperature = temp
	t.Humidity = hum
	t.addRecord()
	var msg = msgUpdate{temp, hum}
	pkt.SetPath("/update").Marshal(&msg).BroadcastUp()
}
