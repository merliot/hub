//go:build prime

package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/prime"
	"github.com/merliot/hub/models/relays"
)

func main() {
	device := prime.New("p1", "prime", "p1")
	server := dean.NewServer(device)
	server.RegisterModel("relays", relays.New)
	server.CreateThing("relays01", "relays", "relays")
	server.Run()
}
