//go:build !rpi && !tinygo

package button

type Button struct {
	Name  string `schema:"desc=Button name"`
	Gpio  string `schema:"desc=GPIO pin"`
	State bool
}

func (b *Button) Setup() error { return nil }
func (b *Button) Get() bool    { return b.State }
