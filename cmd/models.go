package main

import (
	"github.com/merliot/garage"
	"github.com/merliot/hub"
	"github.com/merliot/relays"
)

func registerModels(hub *hub.Hub) {
	hub.RegisterModel("relays", relays.New)
	hub.RegisterModel("garage", garage.New)
}
