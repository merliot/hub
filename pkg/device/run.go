//go:build !tinygo

package device

import "time"

func (d *device) runDemo() {

	// Poll right away, once, and then on ticker
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

func (d *device) startDemo() {
	d.stopChan = make(chan struct{})
	go d.runDemo()
}

func (d *device) stopDemo() {
	close(d.stopChan)
}

// In demo mode, start a go func for each child device
func (d *device) startDemoChildren() {
	for _, childId := range d.Children {
		child := devices[childId]
		child.startDemo()
		child.startDemoChildren()
	}
}

func (d *device) stopDemoChildren() {
	for _, childId := range d.Children {
		child := devices[childId]
		child.stopDemo()
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
