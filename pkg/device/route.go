//go:build !tinygo

package device

var routes map[string]string // key: dst id, value: nexthop id
var routesMu rwMutex

func _routesBuild(parent, base *device) {
	for _, childId := range parent.Children {
		// children point to base
		routes[childId] = base.Id
		child := devices[childId]
		_routesBuild(child, base)
	}
}

func routesBuild(root *device) {

	routesMu.RLock()
	defer routesMu.RUnlock()

	routes = make(map[string]string)

	devicesMu.RLock()
	defer devicesMu.RUnlock()

	// root points to self
	routes[root.Id] = root.Id

	for _, childId := range root.Children {
		// children of root point to self
		routes[childId] = childId
		child := devices[childId]
		_routesBuild(child, child)
	}

	LogInfo("Routes", "map[dst]nexthop", routes)
}

func downlinksRoute(p *Packet) {
	routesMu.RLock()
	nexthop := routes[p.Dst]
	routesMu.RUnlock()
	deviceRouteDown(nexthop, p)
}
