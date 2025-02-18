//go:build !tinygo

package device

import "sync"

type routesJSON map[string]string

type routeMap struct {
	m sync.Map // key: dst device id, value: nexthop device id
}

func (rm *routeMap) store(dst, nexthop string) {
	rm.m.Store(dst, nexthop)
}

func (rm *routeMap) load(dst string) (string, bool) {
	value, ok := rm.m.Load(dst)
	if !ok {
		return nil, false
	}
	return value.(string), true
}

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

func (rm *routeMap) routesBuild() {

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
