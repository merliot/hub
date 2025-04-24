package device

import (
	"net"
	"sync"
)

type linker interface {
	Send(pkt *Packet) error
	RemoteAddr() net.Addr
	Close()
}

type linksJSON []net.Addr

type uplinkMap struct {
	sync.Map // key: linker
}

func (ul *uplinkMap) add(l linker) {
	ul.Store(l, true)
}

func (ul *uplinkMap) remove(l linker) {
	ul.Delete(l)
}

func (ul *uplinkMap) routeAll(pkt *Packet) {
	ul.Range(func(key, value any) bool {
		l := key.(linker)
		l.Send(pkt)
		return true
	})
}

func getLinks(links sync.Map) linksJSON {
	var addrs linksJSON
	links.Range(func(key, value any) bool {
		l := key.(linker)
		addrs = append(addrs, l.RemoteAddr())
		return true
	})
	return addrs
}

func (ul *uplinkMap) getJSON() linksJSON {
	return getLinks(ul.Map)
}

type downlinkMap struct {
	sync.Map // key: device id, value: linker
}

func (dl *downlinkMap) add(id string, l linker) {
	dl.Store(id, l)
}

func (dl *downlinkMap) remove(id string) {
	dl.Delete(id)
}

func (dl *downlinkMap) route(id string, pkt *Packet) {
	if l, ok := dl.Load(id); ok {
		l.(linker).Send(pkt)
	}
}

func (dl *downlinkMap) linkClose(id string) {
	if l, ok := dl.Load(id); ok {
		l.(linker).Close()
	}
}

func (dl *downlinkMap) getJSON() linksJSON {
	return getLinks(dl.Map)
}
