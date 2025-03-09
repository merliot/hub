//go:build !tinygo

package gpio

type Gpio struct {
}

func (g *Gpio) Setup(gpio string) error { return nil }
func (g *Gpio) On()                     {}
func (g *Gpio) Off()                    {}
