package common

import (
	"github.com/merliot/dean"
)

type Common struct {
	dean.Thing
	commonOS
}

func New(id, model, name string) dean.Thinger {
	println("NEW COMMON")
	c := &Common{}
	c.Thing = dean.NewThing(id, model, name)
	c.commonOSInit()
	return c
}
