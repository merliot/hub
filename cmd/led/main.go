package main

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/id"
	"github.com/merliot/hub/models/led"
)

func main() {
	id := id.MAC()
	thing := led.New(id, "led", "led_"+id)
	server := dean.NewServer(thing)
	server.Addr = ":8005"
	server.DialWebSocket("user", "passwd", "wss://hub.merliot.net/ws/1500", thing.Announce())
	//server.DialWebSocket("user", "passwd", "ws://192.168.1.213:8000/ws/1500", thing.Announce())
	go server.ListenAndServe()
	server.Run()
}
