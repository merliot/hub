//go:build rpi

package device

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/merliot/hub/pkg/target"
	"gobot.io/x/gobot/v2/drivers/gpio"
)

// failSafe by turning off all gpios
func failSafe() {
	for _, pin := range target.AllTargets["rpi"].GpioPins {
		rpin := strconv.Itoa(int(pin))
		driver := gpio.NewDirectPinDriver(target.GetAdaptor(), rpin)
		driver.Start()
		driver.Off()
	}

}

func (d *device) runPolling(pollFunc func(pkt *Packet)) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	// Poll right away and then on ticker
	d.stateMu.Lock()
	pollFunc(&Packet{Dst: d.Id})
	d.stateMu.Unlock()

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c:
			failSafe()
			return
		case <-ticker.C:
			d.stateMu.Lock()
			pollFunc(&Packet{Dst: d.Id})
			d.stateMu.Unlock()
		}
	}
}
