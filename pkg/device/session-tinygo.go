//go:build tinygo

package device

type sessionMap struct{}

func (sm *sessionMap) routeAll(pkt *Packet) (err error) {
	return nil
}
