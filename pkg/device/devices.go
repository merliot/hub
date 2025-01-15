package device

import "fmt"

type deviceMap map[string]*device // key: device id

var devices = make(deviceMap)
var devicesMu rwMutex

func deviceNotFound(id string) error {
	return fmt.Errorf("Device '%s' not found", id)
}

func getDevice(id string) (*device, error) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	if d, ok := devices[id]; ok {
		return d, nil
	}
	return nil, deviceNotFound(id)
}
