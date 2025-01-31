package device

import "sync"

type linker interface {
	Send(pkt *Packet) error
	Close()
}

var uplinks sync.Map // keyed by linker

func uplinksAdd(l linker) {
	uplinks.Store(l, true)
}

func uplinksRemove(l linker) {
	uplinks.Delete(l)
}

func uplinksRoute(pkt *Packet) {
	uplinks.Range(func(key, value interface{}) bool {
		ul := key.(linker)
		ul.Send(pkt)
		return true
	})
}

var downlinks sync.Map // keyed by device id

func downlinksAdd(id string, l linker) {
	downlinks.Store(id, l)
}

func downlinksRemove(id string) {
	downlinks.Delete(id)
}

func downlinkRoute(id string, pkt *Packet) {
	if dl, ok := downlinks.Load(id); ok {
		dl.(linker).Send(pkt)
	}
}

func downlinkClose(id string) {
	if dl, ok := downlinks.Load(id); ok {
		dl.(linker).Close()
	}
}
