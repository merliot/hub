//go:build tinygo

package temp

import (
	"errors"
	"fmt"
	"machine"

	"github.com/merliot/hub/pkg/target"
	"tinygo.org/x/drivers/bme280"
	"tinygo.org/x/drivers/dht"
)

const (
	DHT11 uint = iota
	DHT22
	BME280
)

type Temp struct {
	sensor uint
	pin    machine.Pin
	dht    dht.Device
	bme    bme280.Device
}

func (t *Temp) Setup(sensor, gpio string) error {
	switch sensor {
	case "DHT11":
		t.sensor = DHT11
		return t.dhtSetup(gpio)
	case "DHT22":
		t.sensor = DHT22
		return t.dhtSetup(gpio)
	case "BME280":
		t.sensor = BME280
		return t.bmeSetup()
	}
	return fmt.Errorf("Sensor %s not supported")
}

func (t *Temp) Read() (temperature, humidity float32, err error) {
	switch t.sensor {
	case DHT11, DHT22:
		return t.dhtRead()
	case BME280:
		return t.bmeRead()
	}
	return 0.0, 0.0, fmt.Errorf("Sensor %s not supported")
}

func (t *Temp) dhtSetup(gpio string) error {
	sensor := dht.DHT11
	switch t.sensor {
	case DHT22:
		sensor = dht.DHT22
	}
	t.pin = machine.NoPin
	if pin, ok := target.Pin(gpio); ok {
		t.pin = machine.Pin(pin)
		t.dht = dht.New(t.pin, sensor)
		return nil
	}
	return errors.New("Gpio not valid for target")
}

func (t *Temp) dhtRead() (temperature, humidity float32, err error) {
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

func (t *Temp) bmeSetup() error {
	machine.I2C0.Configure(machine.I2CConfig{})
	t.bme = bme280.New(machine.I2C0)
	t.bme.Configure()
	connected := t.bme.Connected()
	if !connected {
		return errors.New("BME280 not detected")
	}
	return nil
}

func (t *Temp) bmeRead() (temperature, humidity float32, err error) {
	temp, err := t.bme.ReadTemperature()
	if err != nil {
		return 0.0, 0.0, err
	}
	hum, err := t.bme.ReadHumidity()
	return float32(temp) / 1000.0, float32(hum) / 100.0, err
}
