package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/modela"
	"github.com/merliot/sw-poc/id"
)

func main() {
	id := id.MAC()
	thing := modela.New(id, "modela", "modela_" + id)
	server := dean.NewServer(thing)
	server.Addr = ":8001"
	server.DialWebSocket("user", "passwd", "wss://sw-poc.merliot.net/ws/1500", thing.Announce())
	go server.ListenAndServe()
	server.Run()
}
