package hub

type handler interface {
	gen() any
	cb(pkt *Packet)
}

// Handler for message type T
type Handler[T any] struct {
	// Callback is called with packet containing type T message
	Callback func(pkt *Packet)
}

// gen an instance of T
func (h *Handler[T]) gen() any {
	var v T
	return &v
}

// cb handles the packet
func (h *Handler[T]) cb(pkt *Packet) {
	if h.Callback != nil {
		h.Callback(pkt)
	}
}

// Handlers is a map of Handlers, keyed by path.
type Handlers map[string]handler

func (d *device) handle(pkt *Packet) {
	d.Lock()
	defer d.Unlock()
	if d.IsSet(flagOnline) {
		if handler, ok := d.Handlers[pkt.Path]; ok {
			LogInfo("Handling", "pkt", pkt)
			handler.cb(pkt)
		}
	}
}
