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
	StartUTC  string
	StopUTC   string
	Gpio      string
	On        bool
	gpio      gpio.Gpio
	startUTC  time.Time
	stopUTC   time.Time
}

func NewModel() device.Devicer {
	return &timer{}
}

func utc() time.Time {
	now := time.Now().UTC()
	hhmmss := fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
	t, _ := time.Parse("15:04:05", hhmmss)
	return t
}

func (t *timer) timeBetween() bool {

	// Get current time of day UTC
	currentTime := utc()

	// Check if current time is between start and stop times
	if t.startUTC.After(t.stopUTC) {
		return currentTime.After(t.startUTC) || currentTime.Before(t.stopUTC)
	} else {
		return currentTime.After(t.startUTC) && currentTime.Before(t.stopUTC)
	}
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
		pkt.SetPath("update").Marshal(&msg).BroadcastUp()
	}
}

func (t *timer) off(pkt *device.Packet) {
	if t.On {
		t.On = false
		t.gpio.Off()
		var msg = msgUpdate{false}
		pkt.SetPath("update").Marshal(&msg).BroadcastUp()
	}
}

func (t *timer) Setup() (err error) {

	t.startUTC, err = time.Parse("15:04", t.StartUTC)
	if err != nil {
		return
	}

	t.stopUTC, err = time.Parse("15:04", t.StopUTC)
	if err != nil {
		return
	}

	// Set system time using NTP protocol
	if err = ntp.SetSystemTime(); err != nil {
		return
	}

	if err = t.gpio.Setup(t.Gpio); err != nil {
		return
	}

	if t.timeBetween() {
		t.On = true
		t.gpio.On()
	}

	return
}

func (t *timer) Poll(pkt *device.Packet) {
	if t.timeBetween() {
		t.on(pkt)
	} else {
		t.off(pkt)
	}
}

func (t *timer) DemoSetup() error {
	if t.StartUTC == "" {
		t.StartUTC = "00:00"
	}
	if t.StopUTC == "" {
		t.StopUTC = "00:00"
	}
	return t.Setup()
}

func (t *timer) DemoPoll(pkt *device.Packet) { t.Poll(pkt) }
