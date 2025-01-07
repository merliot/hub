//go:build !tinygo

package device

var routes map[string]string // key: dst id, value: nexthop id
var routesMu rwMutex

func _routesBuild(parent, base *device) {
	parent.RLock()
	defer parent.RUnlock()

	for _, childId := range parent.Children {
		// children point to base
		routes[childId] = base.Id
		child := devices[childId]
		_routesBuild(child, base)
	}
}

func routesBuild(root *device) {

	routesMu.Lock()
	defer routesMu.Unlock()

	routes = make(map[string]string)

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	root.RLock()
	defer root.RUnlock()

	// root points to self
	routes[root.Id] = root.Id

	for _, childId := range root.Children {
		// children of root point to self
		routes[childId] = childId
		child := devices[childId]
		child.RLock()
		_routesBuild(child, child)
		child.RUnlock()
	}

	LogInfo("Routes", "map[dst]nexthop", routes)
}

func downlinksRoute(p *Packet) {
	routesMu.RLock()
	nexthop := routes[p.Dst]
	routesMu.RUnlock()
	deviceRouteDown(nexthop, p)
}
