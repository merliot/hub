package main

import (
	"github.com/merliot/garage"
	"github.com/merliot/hp2430n"
	"github.com/merliot/hub"
	"github.com/merliot/ps30m"
	"github.com/merliot/relays"
)

func registerModels(hub *hub.Hub) {
	hub.RegisterModel("relays", relays.New)
	hub.RegisterModel("garage", garage.New)
	hub.RegisterModel("hp2430n", hp2430n.New)
	hub.RegisterModel("ps30m", ps30m.New)
}
