// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/hub/devices/camera"
	"github.com/merliot/hub/devices/gadget"
	"github.com/merliot/hub/devices/gps"
	"github.com/merliot/hub/devices/hubdevice"
	"github.com/merliot/hub/devices/locker"
	"github.com/merliot/hub/devices/prostar"
	"github.com/merliot/hub/devices/qrcode"
	"github.com/merliot/hub/devices/relays"
	"github.com/merliot/hub/devices/temp"
	"github.com/merliot/hub/devices/timer"
	"github.com/merliot/hub/pkg/device"
)

var AllModels = device.Models{
	"camera":  &Camera,
	"gadget":  &Gadget,
	"gps":     &Gps,
	"hub":     &Hub,
	"locker":  &Locker,
	"prostar": &Prostar,
	"qrcode":  &Qrcode,
	"relays":  &Relays,
	"temp":    &Temp,
	"timer":   &Timer,
}
var Camera = device.Model{
	Package: "github.com/merliot/hub/devices/camera",
	Maker:   camera.NewModel,
}
var Gadget = device.Model{
	Package: "github.com/merliot/hub/devices/gadget",
	Maker:   gadget.NewModel,
}
var Gps = device.Model{
	Package: "github.com/merliot/hub/devices/gps",
	Maker:   gps.NewModel,
}
var Hub = device.Model{
	Package: "github.com/merliot/hub/devices/hubdevice",
	Maker:   hubdevice.NewModel,
}
var Locker = device.Model{
	Package: "github.com/merliot/hub/devices/locker",
	Maker:   locker.NewModel,
}
var Prostar = device.Model{
	Package: "github.com/merliot/hub/devices/prostar",
	Maker:   prostar.NewModel,
}
var Qrcode = device.Model{
	Package: "github.com/merliot/hub/devices/qrcode",
	Maker:   qrcode.NewModel,
}
var Relays = device.Model{
	Package: "github.com/merliot/hub/devices/relays",
	Maker:   relays.NewModel,
}
var Temp = device.Model{
	Package: "github.com/merliot/hub/devices/temp",
	Maker:   temp.NewModel,
}
var Timer = device.Model{
	Package: "github.com/merliot/hub/devices/timer",
	Maker:   timer.NewModel,
}
