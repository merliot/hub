package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/id"
	"github.com/merliot/hub/models/sense"
)

func main() {
	id := id.MAC()
	thing := sense.New(id, "sense", "sense_"+id)
	server := dean.NewServer(thing)
	server.Addr = ":8003"
	server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/1500", thing.Announce())
	go server.ListenAndServe()
	server.Run()
}
