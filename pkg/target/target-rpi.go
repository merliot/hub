//go:build rpi

package target

import (
	"gobot.io/x/gobot/v2/platforms/raspi"
)

var adaptor *raspi.Adaptor = raspi.NewAdaptor()
var adaptorConnected bool

func GetAdaptor() *raspi.Adaptor {
	if !adaptorConnected {
		if err := adaptor.Connect(); err != nil {
			return nil
		}
		adaptorConnected = true
	}
	return adaptor
}

func Pin(pin string) (GpioPin, bool) {
	gpio, ok := AllTargets["rpi"].GpioPins[pin]
	return gpio, ok
}
