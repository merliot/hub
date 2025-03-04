//go:build tinygo

package gpio

import (
	"errors"
	"machine"

	"github.com/merliot/hub/pkg/target"
)

type Gpio struct {
	pin machine.Pin
}

func (g *Gpio) Setup(gpio string) error {
	g.pin = machine.NoPin
	if pin, ok := target.Pin(gpio); ok {
		g.pin = machine.Pin(pin)
		g.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		return nil
	}
	return errors.New("Gpio not valid for target")
}

func (g *Gpio) On() {
	if g.pin != machine.NoPin {
		g.pin.High()
	}
}

func (g *Gpio) Off() {
	if g.pin != machine.NoPin {
		g.pin.Low()
	}
}
