package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/hub/models/ps30m"
)

var (
	ssid string
	pass string
)

func main() {
	tinynet.NetConnect(ssid, pass)

	thing := ps30m.New("p1", "ps30m", "p1").(*ps30m.Ps30m)

	thing.SetDeployParams("target=nano-rp2040")

	runner := dean.NewRunner(thing)

	runner.DialWebSocket("", "", "ws://192.168.1.213:8000/ws/", thing.Announce())

	runner.Run()
}
