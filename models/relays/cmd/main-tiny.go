//go:build tinygo

package main

import (
	"github.com/meriot/dean"
	"github.com/merliot/hub/models/relays"
)

var (
	ssid string
	pass string
)

func main() {
	tinynet.NetConnect(ssid, pass)
	device := relays.New("relays1", "relays", "relays")
	runner := dean.NewServer(device)
	runner.DialWebSocket()
	runner.Run()
}
