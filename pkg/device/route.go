//go:build !tinygo

package device

type routesJSON map[string]string

func (s *server) routeDown(pkt *Packet) error {

	d, ok := s.devices.load(pkt.Dst)
	if !ok {
		return deviceNotFound(pkt.Dst)
	}

	nexthop := d.nexthop
	if nexthop.isSet(flagMetal) {
		nexthop.handle(pkt)
	} else {
		s.downlinks.route(nexthop.Id, pkt)
	}

	return nil
}
