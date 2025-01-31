//go:build !tinygo

package device

import "sync"

var routes sync.Map // key: dst id, value: nexthop id

func _routesBuild(parent, base *device) {
	for _, childId := range parent.Children {
		// children point to base
		routes.Store(childId, base.Id)
		child := devices[childId]
		child.RLock()
		_routesBuild(child, base)
		child.RUnlock()
	}
}

func routesBuild(root *device) {

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	root.RLock()
	defer root.RUnlock()

	// root points to self
	routes.Store(root.Id, root.Id)

	for _, childId := range root.Children {
		// children of root point to self
		routes.Store(childId, childId)
		child := devices[childId]
		child.RLock()
		_routesBuild(child, child)
		child.RUnlock()
	}

	// Convert sync.Map to regular map for logging
	routeMap := make(map[string]string)
	routes.Range(func(key, value interface{}) bool {
		routeMap[key.(string)] = value.(string)
		return true
	})
	LogDebug("Routes", "map[dst]nexthop", routeMap)
}

func downlinksRoute(p *Packet) {
	if nexthop, ok := routes.Load(p.Dst); ok {
		deviceRouteDown(nexthop.(string), p)
	}
}
