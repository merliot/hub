//go:build tinygo

package led

import "machine"

type Led struct {
	pin machine.Pin
}

func (l *Led) Setup() error {
	l.pin = machine.LED
	l.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return nil
}

func (l *Led) On() {
	l.pin.High()
}

func (l *Led) Off() {
	l.pin.Low()
}
