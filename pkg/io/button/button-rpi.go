//go:build rpi

package button

import (
	"fmt"
	"strconv"
	"time"

	"github.com/merliot/hub/pkg/target"
	"gobot.io/x/gobot/v2/drivers/gpio"
)

const debounceDelay = 50 * time.Millisecond

type Button struct {
	Name             string
	Gpio             string
	State            bool
	driver           *gpio.ButtonDriver
	lastReading      bool
	lastDebounceTime time.Time
}

func (b *Button) Setup() error {
	if b.Gpio == "" {
		return nil
	}
	if pin, ok := target.Pin(b.Gpio); ok {
		spin := strconv.Itoa(int(pin))
		b.driver = gpio.NewButtonDriver(target.GetAdaptor(), spin)
		b.driver.Start()
		return nil
	}
	return fmt.Errorf("No pin for GPIO %s", b.Gpio)
}

func (b *Button) Get() bool {
	if b.driver == nil {
		return false
	}

	reading := b.driver.Active()

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
