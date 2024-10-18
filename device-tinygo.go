//go:build tinygo

package hub

import "fmt"

type deviceOS struct{}

func (d *device) buildOS() error { return nil }

func devicesSendState(l linker) {
	var pkt = &Packet{
		Dst:  root.Id,
		Path: "/state",
	}
	root.RLock()
	pkt.Marshal(root.State)
	root.RUnlock()
	fmt.Println("Sending:", pkt)
	l.Send(pkt)
}

func deviceRouteDown(id string, pkt *Packet) {
	root.handle(pkt)
}
