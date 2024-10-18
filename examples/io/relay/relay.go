//go:build !rpi && !tinygo

package relay

type Relay struct {
	Name  string
	Gpio  string
	State bool
}

func (r *Relay) Setup() error   { return nil }
func (r *Relay) Set(state bool) { r.State = state }
