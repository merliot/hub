//go:build tinygo

package temp

import (
	"errors"
	"machine"

	"github.com/merliot/device/target"
	"tinygo.org/x/drivers/dht"
)

type Temp struct {
	pin machine.Pin
	dht dht.Device
}

func (t *Temp) Setup(sensor, gpio string) error {
	s := dht.DHT11
	switch sensor {
	case "DHT22":
		s = dht.DHT22
	}
	t.pin = machine.NoPin
	if pin, ok := target.Pin(gpio); ok {
		t.pin = machine.Pin(pin)
		t.dht = dht.New(t.pin, s)
		return nil
	}
	return errors.New("Gpio not valid for target")
}

func (t *Temp) Read() (temperature, humidity float32, err error) {
	if t.pin == machine.NoPin {
		err = errors.New("Gpio pin not configured")
		return
	}
	var temp int16
	var hum uint16
	temp, hum, err = t.dht.Measurements()
	if err != nil {
		return
	}
	return float32(temp) / 10.0, float32(hum) / 10.0, nil
}
