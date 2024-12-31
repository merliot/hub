//go:build tinygo

package device

import "machine"

type deviceOS struct{}

type APIs struct{}

func (d *device) _buildOS() error { return nil }

func devicesSendState(l linker) {
	var pkt = &Packet{
		Dst:  root.Id,
		Path: "/state",
	}
	root.RLock()
	pkt.Marshal(root.State)
	root.RUnlock()
	LogInfo("Sending", "pkt", pkt)
	l.Send(pkt)
}

func deviceRouteDown(id string, pkt *Packet) {
	root.handle(pkt)
}

func (d *device) handleReboot(pkt *Packet) {
	machine.CPUReset()
}
