//go:build tinygo

package device

func downlinksRoute(pkt *Packet) {
	root.handle(pkt)
}
