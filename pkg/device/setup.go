//go:build !tinygo

package device

// In demo mode, call DemoSetup() on all devices
func (d *device) demoSetup() (err error) {
	d.stateMu.Lock()
	defer d.stateMu.Unlock()
	if err = d.DemoSetup(); err != nil {
		return
	}
	d.children.drange(func(_ string, child *device) bool {
		if err = child.demoSetup(); err != nil {
			return false
		}
		return true
	})
	return
}

func (d *device) setup() error {
	d.stateMu.Lock()
	defer d.stateMu.Unlock()
	return d.Setup()
}
