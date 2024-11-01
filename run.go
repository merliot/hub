//go:build !tinygo

package hub

import "time"

func (d *device) runDemo() {
	d.stopChan = make(chan struct{})

	// Poll right away and then on ticker
	lockId, err := d.Lock()
	if err != nil {
		panic(err)
	}
	d.DemoPoll(&Packet{Dst: d.Id})
	d.Unlock(lockId)

	ticker := time.NewTicker(d.PollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lockId, err := d.Lock()
			if err != nil {
				panic(err)
			}
			d.DemoPoll(&Packet{Dst: d.Id})
			d.Unlock(lockId)
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
