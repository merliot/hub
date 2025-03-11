//go:build tinygo

package timer

import (
	"time"

	"github.com/merliot/hub/pkg/device"
)

func (t *timer) GetConfig() device.Config {
	return device.Config{
		Model:      "timer",
		State:      t,
		PollPeriod: time.Second,
	}
}
