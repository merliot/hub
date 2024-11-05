//go:build tinygo

package temp

import (
	"github.com/merliot/hub"
)

func (t *temp) GetConfig() hub.Config {
	return hub.Config{
		Model:      "temp",
		State:      t,
		PollPeriod: pollPeriod,
	}
}
