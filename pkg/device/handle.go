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

// PacketHandlers is a map of PacketHandlers
type PacketHandlers map[string]packetHandler // key: path

func (d *device) handle(pkt *Packet) {
	if d.isSet(flagOnline) || pkt.Path == "/online" || pkt.Path == "/downloaded" {
		if handler, ok := d.PacketHandlers[pkt.Path]; ok {
			d.server.logDebug("Handling", "pkt", pkt)
			d.stateMu.Lock()
			handler.cb(pkt)
			d.stateMu.Unlock()
		}
	}
}
