// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/devices/gadget"
	"github.com/merliot/hub/devices/gps"
	"github.com/merliot/hub/devices/hubdevice"
	"github.com/merliot/hub/devices/locker"
	"github.com/merliot/hub/devices/prostar"
	"github.com/merliot/hub/devices/qrcode"
	"github.com/merliot/hub/devices/relays"
	"github.com/merliot/hub/devices/temp"
)

var AllModels = hub.ModelMap{
	"gadget": Gadget,
	"gps": Gps,
	"hub": Hub,
	"locker": Locker,
	"prostar": Prostar,
	"qrcode": Qrcode,
	"relays": Relays,
	"temp": Temp,
}
var Gadget = hub.Model{
	Package: "github.com/merliot/hub/devices/gadget",
	Source: "https://github.com/merliot/hub/tree/main/devices/gadget",
	Maker: gadget.NewModel,
}
var Gps = hub.Model{
	Package: "github.com/merliot/hub/devices/gps",
	Source: "https://github.com/merliot/hub/tree/main/devices/gps",
	Maker: gps.NewModel,
}
var Hub = hub.Model{
	Package: "github.com/merliot/hub/devices/hubdevice",
	Source: "https://github.com/merliot/hub/tree/main/devices/hubdevice",
	Maker: hubdevice.NewModel,
}
var Locker = hub.Model{
	Package: "github.com/merliot/hub/devices/locker",
	Source: "https://github.com/merliot/hub/tree/main/devices/locker",
	Maker: locker.NewModel,
}
var Prostar = hub.Model{
	Package: "github.com/merliot/hub/devices/prostar",
	Source: "https://github.com/merliot/hub/tree/main/devices/prostar",
	Maker: prostar.NewModel,
}
var Qrcode = hub.Model{
	Package: "github.com/merliot/hub/devices/qrcode",
	Source: "https://github.com/merliot/hub/tree/main/devices/qrcode",
	Maker: qrcode.NewModel,
}
var Relays = hub.Model{
	Package: "github.com/merliot/hub/devices/relays",
	Source: "https://github.com/merliot/hub/tree/main/devices/relays",
	Maker: relays.NewModel,
}
var Temp = hub.Model{
	Package: "github.com/merliot/hub/devices/temp",
	Source: "https://github.com/merliot/hub/tree/main/devices/temp",
	Maker: temp.NewModel,
}
