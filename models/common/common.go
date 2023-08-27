package common

import (
	"github.com/merliot/dean"
)

type Common struct {
	dean.Thing
	Targets `json:"-"`
	commonOS
}

func New(id, model, name string, targets []string) dean.Thinger {
	println("NEW COMMON")
	c := &Common{}
	c.Thing = dean.NewThing(id, model, name)
	c.Targets = makeTargets(targets)
	c.commonOSInit()
	return c
}
