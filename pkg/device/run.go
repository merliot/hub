//go:build !tinygo

package device

import "time"

func (d *device) runDemo() {

	var pkt = d.newPacket()

	// Poll right away, once, and then on ticker
	d.stateMu.Lock()
	d.DemoPoll(pkt)
	d.stateMu.Unlock()

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.stateMu.Lock()
			d.DemoPoll(pkt)
			d.stateMu.Unlock()
		case <-d.stopChan:
			return
		}
	}
}

// In demo mode, start a go func for each child device
func (d *device) startDemo() {
	d.stopChan = make(chan struct{})
	go d.runDemo()
}

func (d *device) stopDemo() {
	close(d.stopChan)
}

func (d *device) startDemoChildren() {
	d.children.drange(func(_ string, child *device) bool {
		child.startDemo()
		child.startDemoChildren()
		return true
	})
}

func (d *device) stopDemoChildren() {
	d.children.drange(func(_ string, child *device) bool {
		child.stopDemo()
		child.stopDemoChildren()
		return true
	})
}

func (d *device) run() {
	if d.isSet(flagDemo) {
		d.startDemoChildren()
		d.runPolling(d.DemoPoll)
		d.stopDemoChildren()
	} else {
		d.runPolling(d.Poll)
	}
}
