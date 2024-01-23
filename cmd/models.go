package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/garage"
	"github.com/merliot/relays"
)

func registerModels(server *dean.Server) {
	server.RegisterModel("relays", relays.New)
	server.RegisterModel("garage", garage.New)
}
