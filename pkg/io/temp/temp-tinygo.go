//go:build tinygo

package temp

import (
	"errors"
	"fmt"
	"machine"

	"tinygo.org/x/drivers/bme280"
	"tinygo.org/x/drivers/dht"
)

const (
	BME280 uint = iota
)

type Temp struct {
	sensor uint
	pin    machine.Pin
	dht    dht.Device
	bme    bme280.Device
}

func (t *Temp) Setup(sensor, gpio string) error {
	switch sensor {
	case "BME280":
		t.sensor = BME280
		return t.bmeSetup()
	}
	return fmt.Errorf("Sensor %s not supported")
}

func (t *Temp) Read() (temperature, humidity float32, err error) {
	switch t.sensor {
	case BME280:
		return t.bmeRead()
	}
	return 0.0, 0.0, fmt.Errorf("Sensor %s not supported")
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
