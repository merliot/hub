//go:build !tinygo

package hub

import "time"

func (d *device) runDemo() {

	// Poll right away and then on ticker
	d.Lock()
	d.DemoPoll(&Packet{Dst: d.Id})
	d.Unlock()

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.Lock()
			d.DemoPoll(&Packet{Dst: d.Id})
			d.Unlock()
		case <-d.stopChan:
			return
		}
	}
}

// In demo mode, start a go func for each child device
func (d *device) startDemoChildren() {
	for _, childId := range d.Children {
		child := devices[childId]
		go child.runDemo()
		child.startDemoChildren()
	}
}

func (d *device) stopDemoChildren() {
	for _, childId := range d.Children {
		child := devices[childId]
		close(child.stopChan)
		child.stopDemoChildren()
	}
}

func (d *device) run() {
	if runningDemo {
		d.startDemoChildren()
		d.runPolling(d.DemoPoll)
		d.stopDemoChildren()
	} else {
		d.runPolling(d.Poll)
	}
}
