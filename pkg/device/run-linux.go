//go:build !rpi && !tinygo

package device

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (d *device) runPolling(pollFunc func(pkt *Packet)) {

	var pkt = &Packet{Dst: d.Id}

	// Catch OS kill signals so we can exit gracefully
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	// Poll right away, once, and then on ticker
	d.stateMu.Lock()
	pollFunc(pkt)
	d.stateMu.Unlock()

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c:
			return
		case <-ticker.C:
			d.stateMu.Lock()
			pollFunc(pkt)
			d.stateMu.Unlock()
		}
	}
}
