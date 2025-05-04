//go:build !rpi && !tinygo

package relay

type Relay struct {
	Name  string `schema:"desc=Relay name"`
	Gpio  string `schema:"desc=GPIO pin"`
	State bool
}

func (r *Relay) Setup() error   { return nil }
func (r *Relay) Set(state bool) { r.State = state }
