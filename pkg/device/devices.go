package device

import (
	"fmt"
	"sync"
)

type devicesJSON map[string]*device

type deviceMap struct {
	sync.Map // key: device id, value: *device
}

func (dm *deviceMap) drange(f func(string, *device) bool) {
	dm.m.Range(func(key, value interface{}) bool {
		id := key.(string)
		d := value.(*device)
		return f(id, d)
	})
}

func (dm *deviceMap) setupAPI() {
	dm.drange(func(id string, d *device) bool {
		d.setupAPI()
		return true
	})
}

func (s *server) loadJSON(devs devicesJSON) error {

	s.devices.Clear()
	s.devices.root = nil

skip:
	for id, d := range devs {

		// Hydrate device
		if err := s.buildDevice(id, d); err != nil {
			LogError("Skipping device", "id", id, "err", err)
			delete(devs, id)
			continue
		}

		// Build family tree
		for _, childId := range d.Children {
			child, ok := devs[childId]
			if !ok {
				LogError("Unknown child id, skipping device",
					"device-id", id, "child-id", childId)
				delete(devs, id)
				continue skip
			}
			child.parent = d
			d.children.store(childId, child)
		}
		s.devices.Store(id, d)
	}

	s.root, err := s.devices.findRoot()
	if err != nil {
		return err
	}

	s.devices.setupAPI()

	return nil
}

func (dm *deviceMap) load(id string) (*device, bool) {
	value, ok := dm.Load(id)
	if !ok {
		return nil, false
	}
	return value.(*device), true
}

func (dm *deviceMap) findRoot() (root *device, err error) {

	var count int

	dm.drange(func(id string, d *device) bool {
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

func (dm *deviceMap) buildTree() (root *device, err error) {

	// Build family tree
	dm.drange(func(id string, d *device) bool {
		d.children.Clear()
		for _, childId := range d.Children {
			child, ok := dm.load[childId]
			if !ok {
				LogError("Unknown child id, skipping device",
					"device-id", id, "child-id", childId)
				dm.Delete(id)
				break
			}
			child.parent = d
			d.children.Store(childId, child)
		}
		return true
	})

	// Find root of family tree
	root, err = dm.findRoot()
	if err != nil {
		return
	}

	// Install nexthop routes

	return
}

func (dm *deviceMap) loadJSON(devs devicesJSON) {
	dm.Clear()
	for id, d := range devs {
		dm.Store(id, d)
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
