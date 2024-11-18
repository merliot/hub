//go:build tinygo

package temp

import (
	"github.com/merliot/hub/pkg/device"
)

func (t *temp) GetConfig() device.Config {
	return device.Config{
		Model:      "temp",
		State:      t,
		PollPeriod: pollPeriod,
	}
}
