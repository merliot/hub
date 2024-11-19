//go:build rpi

package relay

import (
	"fmt"
	"strconv"

	"github.com/merliot/hub/pkg/target"
	"gobot.io/x/gobot/v2/drivers/gpio"
)

type Relay struct {
	Name   string
	Gpio   string
	State  bool
	driver *gpio.RelayDriver
}

func (r *Relay) Setup() error {
	if r.Gpio == "" {
		return nil
	}
	if pin, ok := target.Pin(r.Gpio); ok {
		spin := strconv.Itoa(int(pin))
		r.driver = gpio.NewRelayDriver(target.GetAdaptor(), spin)
		r.driver.Start()
		r.driver.Off()
		return nil
	}
	return fmt.Errorf("No pin for GPIO %s", r.Gpio)
}

func (r *Relay) Set(state bool) {
	if r.driver != nil {
		r.State = state
		if state {
			r.driver.On()
		} else {
			r.driver.Off()
		}
	}
}
