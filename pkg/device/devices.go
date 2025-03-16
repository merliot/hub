//go:build !tinygo

package device

import (
	"fmt"
	"sort"
	"sync"
)

type devicesJSON map[string]*device

type deviceMap struct {
	sync.Map // key: device id, value: *device
}

func (dm *deviceMap) drange(f func(string, *device) bool) {
	dm.Range(func(key, value any) bool {
		id := key.(string)
		d := value.(*device)
		return f(id, d)
	})
}

func (dm *deviceMap) sortedByName(f func(string, *device) error) error {

	var devs []*device

	dm.drange(func(id string, d *device) bool {
		devs = append(devs, d)
		return true
	})

	sort.Slice(devs, func(i, j int) bool {
		return devs[i].Name < devs[j].Name
	})

	for _, d := range devs {
		if err := f(d.Id, d); err != nil {
			return err
		}
	}

	return nil
}

func (dm *deviceMap) get(id string) (*device, bool) {
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

func (d *device) _mapRoutes(parent, base *device) {
	parent.children.drange(func(_ string, child *device) bool {
		// Children point to base
		child.nexthop = base
		d._mapRoutes(child, base)
		return true
	})
}

func (d *device) mapRoutes() {

	// Root points to self
	d.nexthop = d

	d.children.drange(func(_ string, child *device) bool {
		// Children of root point to self
		child.nexthop = child
		d._mapRoutes(child, child)
		return true
	})
}

func (dm *deviceMap) buildTree() (root *device, err error) {

	// Build family tree
	dm.drange(func(id string, d *device) bool {
		d.children.Clear()
		for _, childId := range d.Children {
			child, ok := dm.get(childId)
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

	// Install routes
	root.mapRoutes()

	return
}

func (d *device) copyDevice(from *device) {
	d.Model = from.Model
	d.Name = from.Name
	d.Children = from.Children
	d.DeployParams = from.DeployParams
	d.Config = from.Config
	d.flags = from.flags
}

func (s *server) mergeDevice(id string, anchor, newDevice *device) error {

	if id == anchor.Id {
		anchor.Children = newDevice.Children
		// All we want for the anchor is the new anchor child list
		return nil
	}

	device, exists := s.devices.get(id)
	if exists {
		// Better be a ghost
		if !device.isSet(flagGhost) {
			return fmt.Errorf("Device %s already exists, aborting merge", device)
		}
	} else {
		device = newDevice
	}

	device.copyDevice(newDevice)

	if err := s.buildDevice(id, device); err != nil {
		return err
	}

	device.set(flagLocked)
	device.installAPI()

	if !exists {
		s.deviceInstall(device)
	}

	if s.isSet(flagRunningDemo) {
		if err := device.demoSetup(); err != nil {
			return err
		}
		device.startDemo()
	}

	return nil
}

func (s *server) merge(id string, newDevices deviceMap) error {

	// Swing anchor to existing tree
	anchor, ok := s.devices.get(id)
	if !ok {
		return fmt.Errorf("Anchor device does not exists")
	}

	// Recursively ghost the children of the anchor device in the existing
	// device tree.  The children may be resurrected while merging if they
	// exists in the new device tree.
	anchor.children.drange(func(_ string, child *device) bool {
		child.ghost()
		return true
	})

	// Now merge in the new devices, setting up each device as we go
	newDevices.drange(func(id string, newDevice *device) bool {
		if err := s.mergeDevice(id, anchor, newDevice); err != nil {
			return false
		}
		return true
	})

	anchor.set(flagOnline)

	return nil
}

func (dm *deviceMap) loadJSON(devs devicesJSON) {
	dm.Clear()
	for id, d := range devs {
		dm.Store(id, d)
	}
}

func (dm *deviceMap) getJSON() devicesJSON {
	devs := make(devicesJSON)
	dm.drange(func(id string, d *device) bool {
		// Only get alive devices
		if !d.isSet(flagGhost) {
			devs[id] = d
		}
		return true
	})
	return devs
}

func (dm *deviceMap) getRoutes() routesJSON {
	routes := make(routesJSON)
	dm.drange(func(id string, d *device) bool {
		routes[id] = d.nexthop.Id
		return true
	})
	return routes
}

func (dm *deviceMap) sortedId() []string {
	keys := make([]string, 0)
	dm.drange(func(id string, _ *device) bool {
		keys = append(keys, id)
		return true
	})
	sort.Strings(keys)
	return keys
}

type deviceStatus struct {
	Color  string
	Status string
}

func (dm *deviceMap) status() []deviceStatus {
	var statuses = make([]deviceStatus, 0)
	for _, id := range dm.sortedId() {
		d, _ := dm.get(id)
		status := fmt.Sprintf("%-16s %-16s %-16s %3d",
			d.Id, d.Model, d.Name, len(d.Children))
		color := "gold"
		if d.isSet(flagGhost) {
			color = "gray"
		}
		statuses = append(statuses, deviceStatus{
			Color:  color,
			Status: status,
		})
	}
	return statuses
}
