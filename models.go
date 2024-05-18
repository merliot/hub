package hub

import (
	"github.com/merliot/dean"
	"github.com/merliot/garage"
	"github.com/merliot/prostar"
	"github.com/merliot/relays"
	"github.com/merliot/skeleton"
	"github.com/merliot/temp"
)

var models = map[string]dean.ThingMaker{
	"garage":   garage.New,
	"prostar":  prostar.New,
	"relays":   relays.New,
	"skeleton": skeleton.New,
	"temp":     temp.New,
}
