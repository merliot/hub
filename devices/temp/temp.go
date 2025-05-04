package temp

import (
	"time"

	"github.com/merliot/hub/pkg/device"
	io "github.com/merliot/hub/pkg/io/temp"
)

const (
	pollPeriod  = 5 * time.Second
	historyRecs = (60 / 5) * 5 // 5 minute's worth
)

type Record [2]float32

type History [historyRecs]Record

type temp struct {
	Temperature float32 // deg F or C, depends on TempUnits
	Humidity    float32 // %
	Sensor      string  `schema:"desc=Sensor"`
	TempUnits   string  `schema:"desc=Temperature units F or C"`
	History
	io.Temp
}

type msgUpdate struct {
	Temperature float32
	Humidity    float32
}

func NewModel() device.Devicer {
	return &temp{}
}

func (t *temp) addRecord() {
	// Shift all records one position to the left
	for i := 0; i < historyRecs-1; i++ {
		t.History[i] = t.History[i+1]
	}
	// Add the new record at the end
	t.History[historyRecs-1] = Record{t.Temperature, t.Humidity}
}

func (t *temp) update(pkt *device.Packet) {
	t.addRecord()
	pkt.Unmarshal(t).BroadcastUp()
}

func (t *temp) read() error {
	temp, hum, err := t.Read()
	if err != nil {
		return err
	}
	if t.TempUnits == "F" {
		// Convert from Celsius
		temp = (temp * 9.0 / 5.0) + 32.0
	}
	t.Temperature = temp
	t.Humidity = hum
	t.addRecord()
	return nil
}

func (t *temp) Setup() error {
	if err := t.Temp.Setup(t.Sensor, ""); err != nil {
		return err
	}
	return t.read()
}

func (t *temp) Poll(pkt *device.Packet) {
	if err := t.read(); err != nil {
		println("Temp device poll read", "err", err)
		return
	}
	var msg = msgUpdate{t.Temperature, t.Humidity}
	pkt.SetPath("update").Marshal(&msg).BroadcastUp()
}
