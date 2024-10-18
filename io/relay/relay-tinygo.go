//go:build tinygo

package relay

import (
	"machine"

	"github.com/merliot/hub/target"
)

type Relay struct {
	Name  string
	Gpio  string
	State bool
	pin   machine.Pin
}

func (r *Relay) Setup() error {
	r.pin = machine.NoPin
	if pin, ok := target.Pin(r.Gpio); ok {
		r.pin = machine.Pin(pin)
		r.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		r.pin.Low()
	}
	return nil
}

func (r *Relay) Set(state bool) {
	if r.pin != machine.NoPin {
		r.State = state
		if state {
			r.pin.High()
		} else {
			r.pin.Low()
		}
	}
}
