package device

import (
	"fmt"
	"sync"
)

type devicesJSON map[string]*device

type deviceMap struct {
	sync.Map // key: device id, value: *device
	root     *device
}

func (dm *deviceMap) store(id string, d *device) {
	dm.m.Store(id, d)
}

func (dm *deviceMap) load(id string) (*device, bool) {
	value, ok := dm.m.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*device), true
}

func (dm *deviceMap) destroy(id string) {
	dm.m.Delete(id)
}

func (dm *deviceMap) drange(f func(string, *device) bool) {
	dm.m.Range(func(key, value interface{}) bool {
		id := key.(string)
		d := value.(*device)
		return f(id, d)
	})
}

func (s *server) buildDevice(id string, d *device) error {
	if id != d.Id {
		return fmt.Errorf("Mismatching Ids")
	}
	model, ok := s.models[d.Model]
	if !ok {
		return fmt.Errorf("Model '%s' not registered", d.Model)
	}
	return d.build(model.Maker)
}

func (s *server) loadJSON(devs devicesJSON) {
	s.Lock()
	defer s.Unlock()

	s.devices.Clear()
	s.devices.root = nil

	for id, d := range devs {
		if err := s.buildDevice(id, d); err != nil {
			LogError("Skipping device", "id", id, "err", err)
			delete(devs, id)
		}
	}

	for id, d := range devs {
		s.devices.Store(id, d)
	}
}

func (dm *deviceMap) saveJSON() devicesJSON {
	devs := make(devicesJSON)
	dm.drange(func(id string, d *device) bool {
		devs[id] = d
		return true
	})
	return devs
}

func (dm *deviceMap) buildFamilyTree() {

	dm.drange(func(id string, d *device) bool {
		d.Lock()
		defer d.Unlock()
		for _, childId := range d.Children {
			child, ok := dm.load(childId)
			if !ok {
				LogError("Unknown child id, skipping device", "device-id", id, "child-id", childId)
				s.devices.destroy(id)
				return true
			}
			child.parent = d
			d.children.store(childId, child)
		}
		return true
	})
}

func (dm *deviceMap) findRoot() (root *device, err error) {

	var count int

	dm.drange(func(id string, d *device) bool {
		d.Lock()
		defer d.Unlock()
		if d.parent == nil {
			root = d
			count++
		}
		return true
	})

	switch count {
	case 0:
		return nil, fmt.Errorf("No root device found")
	case 1:
		root.set(flagRoot)
		return root, nil
	default:
		return nil, fmt.Errorf("Multiple root devices found")
	}
}

func (dm *deviceMap) setupAPI() {
	dm.drange(func(id string, d *device) bool {
		d.setupAPI()
		return true
	})
}

func (dm *deviceMap) _mapRoutes(parent *device, baseId string) {
	parent.children.drange(func(childId string, child *device) {
		// Children point to base
		dm.routes.store(childId, baseId)
		dm._mapRoutes(child, baseId)
	})
}

func (dm *deviceMap) mapRoutes() {

	dm.routes.Clear()

	// Root points to self
	dm.routes.store(dm.root.Id, dm.root.Id)

	dm.root.children.drange(func(childId string, child *device) bool {
		// Children of root point to self
		dm.routes.store(childId, childId)
		dm._mapRoutes(child, childId)
	})
}

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
