package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/id"
	"github.com/merliot/sw-poc/models/gps"
)

func main() {
	id := id.MAC()
	thing := gps.New(id, "gps", "gps_"+id)
	server := dean.NewServer(thing)
	server.Addr = ":8002"
	server.DialWebSocket("user", "passwd", "wss://sw-poc.merliot.net/ws/1500", thing.Announce())
	go server.ListenAndServe()
	server.Run()
}
