//go:build !tinygo

package device

// In demo mode, call DemoSetup() on all devices
func (d *device) demoSetup() error {
	if err := d.DemoSetup(); err != nil {
		return err
	}
	for _, childId := range d.Children {
		child := devices[childId]
		if err := child.demoSetup(); err != nil {
			return err
		}
	}
	return nil
}

func (d *device) _setup() error {
	if runningDemo {
		return d.demoSetup()
	}
	return d.Setup()
}

func (d *device) setup() error {
	d.Lock()
	defer d.Unlock()
	return d._setup()
}
