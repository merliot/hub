package device

type linker interface {
	Send(pkt *Packet) error
	Close()
}

var uplinks = make(map[linker]bool) // keyed by linker
var uplinksMu rwMutex

func uplinksAdd(l linker) {
	uplinksMu.Lock()
	defer uplinksMu.Unlock()
	uplinks[l] = true
}

func uplinksRemove(l linker) {
	uplinksMu.Lock()
	defer uplinksMu.Unlock()
	delete(uplinks, l)
}

func uplinksRoute(pkt *Packet) {
	uplinksMu.RLock()
	defer uplinksMu.RUnlock()
	for ul := range uplinks {
		ul.Send(pkt)
	}
}

var downlinks = make(map[string]linker) // keyed by device id
var downlinksMu rwMutex

func downlinksAdd(id string, l linker) {
	downlinksMu.Lock()
	defer downlinksMu.Unlock()
	downlinks[id] = l
}

func downlinksRemove(id string) {
	downlinksMu.Lock()
	defer downlinksMu.Unlock()
	delete(downlinks, id)
}

func downlinkRoute(id string, pkt *Packet) {
	downlinksMu.RLock()
	defer downlinksMu.RUnlock()
	if dl, ok := downlinks[id]; ok {
		dl.Send(pkt)
	}
}

func downlinkClose(id string) {
	downlinksMu.RLock()
	defer downlinksMu.RUnlock()
	if dl, ok := downlinks[id]; ok {
		dl.Close()
	}
}
