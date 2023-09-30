//go:build tinygo

package relays

import (
	"machine"

	"github.com/merliot/dean"
)

type relaysOS struct {
}

func (r *Relays) relaysOSNew() {
}

type Relay struct {
	Name  string
	Gpio  string
	State bool
	pin   machine.Pin
}

func (r *Relay) On() {
	if r.pin != machine.NoPin {
		r.pin.High()
	}
}

func (r *Relay) Off() {
	if r.pin != machine.NoPin {
		r.pin.Low()
	}
}

func (r *Relays) runOS(i *dean.Injector) {

	for i := range r.Relays {
		relay := &r.Relays[i]
		relay.pin = machine.NoPin
		if pin, ok := r.pins()[relay.Gpio]; ok {
			relay.pin = machine.Pin(pin)
			relay.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		}
	}

	select {}
}
