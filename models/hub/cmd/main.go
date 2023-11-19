//go:build !tinygo && !prime

package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/gps"
	"github.com/merliot/hub/models/hp2430n"
	"github.com/merliot/hub/models/hub"
	"github.com/merliot/hub/models/move"
	"github.com/merliot/hub/models/ps30m"
	"github.com/merliot/hub/models/relays"
	"github.com/merliot/hub/models/sign"
	"github.com/merliot/hub/models/skeleton"
	"github.com/merliot/hub/models/uv"
)

func main() {
	device := hub.New("h1", "hub", "h1").(*hub.Hub)
	server := dean.NewServer(device)
	device.SetServer(server)

	server.RegisterModel("ps30m", ps30m.New)
	server.RegisterModel("hp2430n", hp2430n.New)
	server.RegisterModel("gps", gps.New)
	server.RegisterModel("relays", relays.New)
	server.RegisterModel("sign", sign.New)
	server.RegisterModel("uv", uv.New)
	server.RegisterModel("move", move.New)
	server.RegisterModel("skeleton", skeleton.New)

	server.Run()
}
