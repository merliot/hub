//go:build !tinygo && !prime

package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/relays"
)

func main() {
	relays := relays.New("relays01", "relays", "relays").(*relays.Relays)
	relays.SetDeployParams("target=nano-rp2040&relay1=kitchen+and+bath&relay2=closet&relay3=hallway&relay4=&gpio1=D2&gpio2=D4&gpio3=D5&gpio4=")
	server := dean.NewServer(relays)
	server.Dial()
	server.Run()
}
