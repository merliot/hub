package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/relays"
)

func main() {
	thing := relays.New("relays01", "relays", "relays").(*relays.Relays)

	demo, _ := os.LookupEnv("DEMO")
	thing.Demo = (demo != "")

	thing.DeployParams = "target=rpi&amp;http=on&amp;relay1=foo&amp;relay2=&amp;relay3=&amp;relay4=&amp;gpio1=31&amp;gpio2=33&amp;gpio3=35&amp;gpio4=37"
	/*
	thing.SetRelay(0, "Kitchen", "31")
	thing.SetRelay(1, "Living Room", "33")
	thing.SetRelay(2, "Bath Room", "35")
	thing.SetRelay(3, "Bed Room", "37")
	*/

	server := dean.NewServer(thing)

	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	server.DialWebSocket(user, passwd, "ws://192.168.1.213:8000/ws/1500", thing.Announce())
	//server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/1500", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
