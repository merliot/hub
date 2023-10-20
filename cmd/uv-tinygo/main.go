package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/hub/models/uv"
)

var (
	ssid string
	pass string
)

func main() {
	tinynet.NetConnect(ssid, pass)

	thing := uv.New("u1", "uv", "nano-uv").(*uv.UV)

	thing.SetDeployParams("target=nano-rp2040")

	runner := dean.NewRunner(thing)

	runner.DialWebSocket("", "", "ws://192.168.1.213:8000/ws/", thing.Announce())

	runner.Run()
}
