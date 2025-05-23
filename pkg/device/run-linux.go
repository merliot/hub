//go:build !rpi && !tinygo

package device

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (d *device) runPolling(pollFunc func(pkt *Packet)) {

	var pkt = d.newPacket()

	d.start()

	// Catch OS kill signals so we can exit gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return
		case <-c:
			return
		case <-ticker.C:
			d.stateMu.Lock()
			pollFunc(pkt)
			d.stateMu.Unlock()
		}
	}
}
