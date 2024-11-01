//go:build !tinygo

package hub

import (
	"github.com/ietxaniz/delock"
)

var routes map[string]string // key: dst id, value: nexthop id
var routesMu delock.RWMutex

func _routesBuild(parent, base *device) {
	for _, childId := range parent.Children {
		// children point to base
		routes[childId] = base.Id
		child := devices[childId]
		_routesBuild(child, base)
	}
}

func routesBuild(root *device) {

	lockId, err := routesMu.RLock()
	if err != nil {
		panic(err)
	}
	defer routesMu.RUnlock(lockId)

	routes = make(map[string]string)

	lockId2, err := devicesMu.RLock()
	if err != nil {
		panic(err)
	}
	defer devicesMu.RUnlock(lockId2)

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
	lockId, err := routesMu.RLock()
	if err != nil {
		panic(err)
	}
	nexthop := routes[p.Dst]
	routesMu.RUnlock(lockId)
	deviceRouteDown(nexthop, p)
}
