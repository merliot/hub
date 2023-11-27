//go:build !tinygo && !prime

package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/relays"
)

func main() {
	device := relays.New("relays1", "relays", "relays")
	server := dean.NewServer(device)
	server.Dial()
	server.Run()
}
