package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/relays"
)

func main() {
	thing := relays.New("relays01", "relays", "relays").(*relays.Relays)

	thing.SetDeployParams("target=nano-rp2040&amp;relay1=&amp;relay2=&amp;relay3=&amp;relay4=&amp;gpio1=D2&amp;gpio2=&amp;gpio3=&amp;gpio4=")

	runner := dean.NewRunner(thing)

	runner.DialWebSocket("", "", "ws://192.168.1.213:8000/ws/1500", thing.Announce())

	runner.Run()
}
