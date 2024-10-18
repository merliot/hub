// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/examples/gadget"
	"github.com/merliot/hub/examples/gps"
	"github.com/merliot/hub/hub"
	"github.com/merliot/hub/examples/relays"
)

var AllModels = hub.ModelMap{
	"gadget": Gadget,
	"gps": Gps,
	"hub": Hub,
	"relays": Relays,
}
var Gadget = hub.Model{
	Package: "github.com/merliot/hub/examples/gadget",
	Source: "https://github.com/merliot/hub/tree/main/examples/gadget",
	Maker: gadget.NewModel,
}
var Gps = hub.Model{
	Package: "github.com/merliot/hub/examples/gps",
	Source: "https://github.com/merliot/hub/tree/main/examples/gps",
	Maker: gps.NewModel,
}
var Hub = hub.Model{
	Package: "github.com/merliot/hub/hub",
	Source: "https://github.com/merliot/hub/tree/main/hub",
	Maker: hub.NewModel,
}
var Relays = hub.Model{
	Package: "github.com/merliot/hub/examples/relays",
	Source: "https://github.com/merliot/hub/tree/main/examples/relays",
	Maker: relays.NewModel,
}
