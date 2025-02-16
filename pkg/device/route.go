//go:build !tinygo

package device

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

func (s *server) routesBuild() {

	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()

	s.root.RLock()
	defer s.root.RUnlock()

	// root points to self
	s.routes.Store(s.root.Id, s.root.Id)

	for _, childId := range s.root.Children {
		// children of root point to self
		s.routes.Store(childId, childId)
		child := s.devices[childId]
		child.RLock()
		_routesBuild(child, child)
		child.RUnlock()
	}

	// Convert sync.Map to regular map for logging
	routeMap := make(map[string]string)
	s.routes.Range(func(key, value interface{}) bool {
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
