package common

import (
	"github.com/merliot/dean"
)

type Common struct {
	dean.Thing
	dean.ThingMsg
	Identity Identity
}

type Identity struct {
	Id    string
	Model string
	Name  string
}

func New(id, model, name string) dean.Thinger {
	println("NEW COMMON")
	return &Common{
		Thing: dean.NewThing(id, model, name),
		Identity: Identity{id, model, name},
	}
}
