// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/examples/gadget"
	"github.com/merliot/hub/examples/gps"
	"github.com/merliot/hub/hubdevice"
	"github.com/merliot/hub/examples/locker"
	"github.com/merliot/hub/examples/relays"
)

var AllModels = hub.ModelMap{
	"gadget": Gadget,
	"gps": Gps,
	"hub": Hub,
	"locker": Locker,
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
	Package: "github.com/merliot/hub/hubdevice",
	Source: "https://github.com/merliot/hub/tree/main/hubdevice",
	Maker: hubdevice.NewModel,
}
var Locker = hub.Model{
	Package: "github.com/merliot/hub/examples/locker",
	Source: "https://github.com/merliot/hub/tree/main/examples/locker",
	Maker: locker.NewModel,
}
var Relays = hub.Model{
	Package: "github.com/merliot/hub/examples/relays",
	Source: "https://github.com/merliot/hub/tree/main/examples/relays",
	Maker: relays.NewModel,
}
