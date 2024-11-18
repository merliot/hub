//go:build !tinygo

package led

type Led struct {
}

func (l *Led) Setup() error { return nil }
func (l *Led) On()          {}
func (l *Led) Off()         {}
