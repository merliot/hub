package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/hub/models/sign"
)

var (
	ssid string
	pass string
)

func main() {
	tinynet.NetConnect(ssid, pass)

	thing := sign.New("s1", "sign", "wio-sign").(*sign.Sign)

	thing.SetDeployParams("target=wioterminal")

	runner := dean.NewRunner(thing)

	runner.DialWebSocket("", "", "ws://192.168.1.213:8000/ws/1500", thing.Announce())

	runner.Run()
}
