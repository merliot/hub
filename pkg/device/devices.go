package device

import "fmt"

type deviceMap map[string]*device // key: device id

func deviceNotFound(id string) error {
	return fmt.Errorf("Device '%s' not found", id)
}

func (s *server) getDevice(id string) (*device, error) {
	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()
	if d, ok := s.devices[id]; ok {
		return d, nil
	}
	return nil, deviceNotFound(id)
}

func (s *server) aliveDevices() (alive deviceMap) {
	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()
	alive = make(deviceMap)
	for id, d := range s.devices {
		if !d.isSet(flagGhost) {
			alive[id] = d
		}
	}
	return
}
