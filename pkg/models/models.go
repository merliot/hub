// This file auto-generated from ./cmd/gen-models using models.json as input

package models

import (
	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/devices/camera"
	"github.com/merliot/hub/devices/hubdevice"
)

var AllModels = device.ModelMap{
	"camera": Camera,
	"hub": Hub,
}
var Camera = device.Model{
	Package: "github.com/merliot/hub/devices/camera",
	Maker: camera.NewModel,
}
var Hub = device.Model{
	Package: "github.com/merliot/hub/devices/hubdevice",
	Maker: hubdevice.NewModel,
}
