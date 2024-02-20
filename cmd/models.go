package main

import (
	"github.com/merliot/garage"
	"github.com/merliot/hub"
	"github.com/merliot/ps30m"
	"github.com/merliot/relays"
	"github.com/merliot/skeleton"
)

func registerModels(hub *hub.Hub) {
	hub.RegisterModel("garage", garage.New)
	hub.RegisterModel("ps30m", ps30m.New)
	hub.RegisterModel("relays", relays.New)
	hub.RegisterModel("skeleton", skeleton.New)
}
