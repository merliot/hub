package buttons

import (
	"math/rand"
	"time"

	"github.com/merliot/hub/pkg/device"
)

func (b *buttons) DemoSetup() error {
	rand.Seed(time.Now().UnixNano())
	return nil
}

func (b *buttons) DemoPoll(pkt *device.Packet) {
	for i := range b.Buttons {
		button := &b.Buttons[i]
		// Want button to flip on average every 10s
		// Probability of flipping: 10ms / 10s = 0.001
		if rand.Float64() < 0.001 {
			button.State = !button.State
			var update = msgUpdate{i, button.State}
			pkt.SetPath("update").Marshal(&update).BroadcastUp()
		}
	}
}
