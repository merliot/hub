//go:build tinygo

package device

import "machine"

type Maker func() Devicer
type Model struct{ Maker }
type APIs map[string]any
type deviceMap struct{}
type deviceOS struct{}

func (s *server) buildOS(d *device) error { return nil }

func (dm *deviceMap) get(id string) (*device, bool) { return nil, false }

func (d *device) handleReboot(pkt *Packet) {
	machine.CPUReset()
}
