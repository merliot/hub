package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/hub/models/hp2430n"
)

var (
	ssid string
	pass string
)

func main() {
	tinynet.NetConnect(ssid, pass)

	thing := hp2430n.New("h2", "hp2430n", "nano-hp2430n").(*hp2430n.Hp2430n)

	thing.SetDeployParams("target=nano-rp2040")

	runner := dean.NewRunner(thing)

	runner.DialWebSocket("", "", "ws://192.168.1.213:8000/ws/", thing.Announce())

	runner.Run()
}
