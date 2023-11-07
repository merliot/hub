package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean/tinynet"
	"github.com/merliot/hub/models/move"
)

var (
	ssid string
	pass string
)

func main() {
	tinynet.NetConnect(ssid, pass)

	thing := move.New("m1", "move", "nano-move").(*move.Move)

	thing.SetDeployParams("target=wioterminal")

	runner := dean.NewRunner(thing)

	runner.DialWebSocket("", "", "ws://192.168.1.213:8000/ws/", thing.Announce())

	runner.Run()
}
