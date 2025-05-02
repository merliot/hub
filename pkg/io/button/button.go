//go:build !rpi && !tinygo

package button

type Button struct {
	Name  string
	Gpio  string
	State bool
}

func (b *Button) Setup() error { return nil }
func (b *Button) Get() bool    { return b.State }
