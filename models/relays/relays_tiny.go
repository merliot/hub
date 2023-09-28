//go:build tinygo

package relays

import (
	"machine"

	"github.com/merliot/dean"
	_ "github.com/merliot/dean/tinynet"
)

type relaysOS struct {
}

func (r *Relays) relaysOSNew() {
}

type Relay struct {
	Name   string
	Gpio   string
	State  bool
	pin    machine.Pin
}

func (r *Relay) On() {
	r.pin.High()
}

func (r *Relay) Off() {
	r.pin.Low()
}

func (r *Relays) runOS(i *dean.Injector) {
	for i, _ := range r.Relays {
		relay := &r.Relays[i]
		if relay.Gpio == "" {
			continue
		}
		if pin, ok := r.pins()[relay.Gpio]; ok {
			relay.pin = machine.Pin(pin)
		}
	}

	select{}
}
