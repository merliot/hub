package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/sense"
	"github.com/merliot/sw-poc/id"
)

func main() {
	id := id.MAC()
	thing := sense.New(id, "sense", "sense_" + id)
	server := dean.NewServer(thing)
	server.Addr = ":8003"
	server.DialWebSocket("user", "passwd", "wss://sw-poc.merliot.net/ws/1500", thing.Announce())
	go server.ListenAndServe()
	server.Run()
}
