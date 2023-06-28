package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/modela"
)

func main() {
	thing := modela.New("modela-01", "modela", "modela-01")
	server := dean.NewServer(thing)
	server.Addr = ":8001"
//	server.DialWebSocket("user", "passwd", "ws://localhost:8000/ws/1500", thing.Announce())
	server.DialWebSocket("user", "passwd", "wss://sw-poc.merliot.net/ws/1500", thing.Announce())
	go server.ListenAndServe()
	server.Run()
}
