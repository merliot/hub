package timer

import (
	"fmt"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/io/gpio"
	"github.com/merliot/hub/pkg/io/ntp"
)

type timer struct {
	StartHHMM string
	StopHHMM  string
	TZ        string
	Gpio      string
	On        bool
	gpio      gpio.Gpio
	startTime time.Time
	stopTime  time.Time
}

func NewModel() device.Devicer {
	t := &timer{
		StartHHMM: "10:00",
		StopHHMM:  "14:00",
	}
	t.TZ, _ = time.Now().Zone()
	return t
}

func (t *timer) timeBetween() bool {

	// Get current time
	currentTime := time.Now()
	currentTime = time.Date(0, 1, 1,
		currentTime.Hour(),
		currentTime.Minute(),
		currentTime.Second(), 0,
		time.UTC)

	// Check if current time is between start and stop times
	if t.startTime.After(t.stopTime) {
		return currentTime.After(t.startTime) || currentTime.Before(t.stopTime)
	} else {
		return currentTime.After(t.startTime) && currentTime.Before(t.stopTime)
	}

	return false
}

type msgUpdate struct {
	On bool
}

func (t *timer) update(pkt *device.Packet) {
	pkt.Unmarshal(t).BroadcastUp()
}

func (t *timer) on(pkt *device.Packet) {
	if !t.On {
		t.On = true
		t.gpio.On()
		var msg = msgUpdate{true}
		pkt.SetPath("/update").Marshal(&msg).BroadcastUp()
		ntp.SetSystemTime()
	}
}

func (t *timer) off(pkt *device.Packet) {
	if t.On {
		t.On = false
		t.gpio.Off()
		var msg = msgUpdate{false}
		pkt.SetPath("/update").Marshal(&msg).BroadcastUp()
		ntp.SetSystemTime()
	}
}

func (t *timer) Setup() (err error) {

	tzOffset, ok := tzMap[t.TZ]
	if !ok {
		return fmt.Errorf("Time zone '%s' not found", t.TZ)
	}

	// Parse start and stop times
	t.startTime, err = time.Parse("15:04 -0700", t.StartHHMM+" "+tzOffset)
	if err != nil {
		return err
	}
	t.stopTime, err = time.Parse("15:04 -0700", t.StopHHMM+" "+tzOffset)
	if err != nil {
		return err
	}

	// Set system time using NTP protocol
	if err = ntp.SetSystemTime(); err != nil {
		return err
	}

	if err = t.gpio.Setup(t.Gpio); err != nil {
		return err
	}
	if t.timeBetween() {
		t.On = true
		t.gpio.On()
	}

	return nil
}

func (t *timer) Poll(pkt *device.Packet) {
	if t.timeBetween() {
		t.on(pkt)
	} else {
		t.off(pkt)
	}
}

func (t *timer) DemoSetup() error            { return t.Setup() }
func (t *timer) DemoPoll(pkt *device.Packet) { t.Poll(pkt) }
