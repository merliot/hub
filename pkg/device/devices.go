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
	root.mapRoutes()

	return
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

func deviceNotFound(id string) error {
	return fmt.Errorf("Device '%s' not found", id)
}

func (dm *deviceMap) sortedId() []string {
	keys := make([]string, 0, len(devices))
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
	var statuses = make([]deviceStatus)
	for _, id := range dm.sortedId() {
		d, _ := dm.load(id)
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
