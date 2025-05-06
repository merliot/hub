//go:build tinygo

package button

import (
	"fmt"
	"machine"
	"time"

	"github.com/merliot/hub/pkg/target"
)

const debounceDelay = 50 * time.Millisecond

type Button struct {
	Name             string
	Gpio             string
	State            bool
	pin              machine.Pin
	lastReading      bool
	lastDebounceTime time.Time
}

func (b *Button) Setup() error {
	if b.Gpio == "" {
		return nil
	}
	if pin, ok := target.Pin(b.Gpio); ok {
		b.pin = machine.Pin(pin)
		b.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		return nil
	}
	return fmt.Errorf("No pin for GPIO %s", b.Gpio)
}

func (b *Button) Get() bool {
	if b.Gpio == "" {
		return false
	}

	reading := !b.pin.Get() // Active low

	// If the switch changed, due to noise or pressing:
	if reading != b.lastReading {
		// Reset the debouncing timer
		b.lastDebounceTime = time.Now()
	}

	// Check if enough time has passed since last state change
	if time.Since(b.lastDebounceTime) > debounceDelay {
		// If the button state has changed
		if reading != b.State {
			b.State = reading
		}
	}

	b.lastReading = reading

	return b.State
}
