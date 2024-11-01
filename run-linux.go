//go:build !rpi && !tinygo

package hub

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (d *device) runPolling(pollFunc func(pkt *Packet)) {

	// Catch OS kill signals so we can exit gracefully
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	// Poll right away and then on ticker
	lockId, err := d.Lock()
	if err != nil {
		panic(err)
	}
	pollFunc(&Packet{Dst: d.Id})
	d.Unlock(lockId)

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c:
			return
		case <-ticker.C:
			lockId2, err := d.Lock()
			if err != nil {
				panic(err)
			}
			pollFunc(&Packet{Dst: d.Id})
			d.Unlock(lockId2)
		}
	}
}
