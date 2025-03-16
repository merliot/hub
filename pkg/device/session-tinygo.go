//go:build tinygo

package device

import "errors"

type sessionMap struct{}

func (sm *sessionMap) routeAll(pkt *Packet) (err error) {
	return errors.New("Sessions.routeAll not implemented")
}
