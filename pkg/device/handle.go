package device

type packetHandler interface {
	gen() any
	cb(pkt *Packet)
}

// PacketHandler for message type T
type PacketHandler[T any] struct {
	// Callback is called with packet containing type T message
	Callback func(pkt *Packet)
}

// gen an instance of T
func (h *PacketHandler[T]) gen() any {
	var v T
	return &v
}

// cb handles the packet
func (h *PacketHandler[T]) cb(pkt *Packet) {
	if h.Callback != nil {
		h.Callback(pkt)
	}
}

// PacketHandlers is a map of Handlers, keyed by path.
type PacketHandlers map[string]packetHandler

func (d *device) handle(pkt *Packet) {
	if d.isSet(flagOnline) || pkt.Path == "/online" || pkt.Path == "/downloaded" {
		if handler, ok := d.PacketHandlers[pkt.Path]; ok {
			LogDebug("Handling", "pkt", pkt)
			d.stateMu.Lock()
			handler.cb(pkt)
			d.stateMu.Unlock()
		}
	}
}

func (s *server) handle(pkt *Packet) error {
	if pkt.Dst == "" {
		// Run server handler
		if handler, ok := s.packetHandlers[pkt.Path]; ok {
			LogDebug("Handling", "pkt", pkt)
			handler.cb(pkt)
		}
		return nil
	}

	d, ok := s.devices.load(pkt.Dst)
	if !ok {
		return deviceNotFound(pkt.Dst)
	}

	// Run device handler
	d.handle(pkt)
	return nil
}
