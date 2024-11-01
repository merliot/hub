package hub

import "github.com/ietxaniz/delock"

type linker interface {
	Send(pkt *Packet) error
	Close()
}

var uplinks = make(map[linker]bool) // keyed by linker
var uplinksMu delock.RWMutex

func uplinksAdd(l linker) {
	lockId, err := uplinksMu.Lock()
	if err != nil {
		panic(err)
	}
	defer uplinksMu.Unlock(lockId)
	uplinks[l] = true
}

func uplinksRemove(l linker) {
	lockId, err := uplinksMu.Lock()
	if err != nil {
		panic(err)
	}
	defer uplinksMu.Unlock(lockId)
	delete(uplinks, l)
}

func uplinksRoute(pkt *Packet) {
	lockId, err := uplinksMu.RLock()
	if err != nil {
		panic(err)
	}
	defer uplinksMu.RUnlock(lockId)
	for ul := range uplinks {
		ul.Send(pkt)
	}
}

var downlinks = make(map[string]linker) // keyed by device id
var downlinksMu delock.RWMutex

func downlinksAdd(id string, l linker) {
	lockId, err := downlinksMu.Lock()
	if err != nil {
		panic(err)
	}
	defer downlinksMu.Unlock(lockId)
	downlinks[id] = l
}

func downlinksRemove(id string) {
	lockId, err := downlinksMu.Lock()
	if err != nil {
		panic(err)
	}
	defer downlinksMu.Unlock(lockId)
	delete(downlinks, id)
}

func downlinkRoute(pkt *Packet) {
	lockId, err := downlinksMu.RLock()
	if err != nil {
		panic(err)
	}
	defer downlinksMu.RUnlock(lockId)
	if dl, ok := downlinks[pkt.Dst]; ok {
		dl.Send(pkt)
	}
}

func downlinkClose(id string) {
	lockId, err := downlinksMu.RLock()
	if err != nil {
		panic(err)
	}
	defer downlinksMu.RUnlock(lockId)
	if dl, ok := downlinks[id]; ok {
		dl.Close()
	}
}
