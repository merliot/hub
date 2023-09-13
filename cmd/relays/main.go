package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/relays"
)

func main() {
	thing := relays.New("relays01", "relays", "relays").(*relays.Relays)

	thing.Demo()
	thing.SetRelay(1, "Kitchen", "32")
	thing.SetRelay(2, "Living Room", "33")

	server := dean.NewServer(thing)

	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	server.DialWebSocket(user, passwd, "ws://127.0.0.1:8000/ws/1500", thing.Announce())
	//server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/1500", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
