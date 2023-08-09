package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/ps30m"
)

func main() {
	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")
	demo, _ := os.LookupEnv("DEMO")

	thing := ps30m.New("230042", "ps30m", "poc1").(*ps30m.Ps30m)

	if demo != "" {
		thing.Demo()
	}

	server := dean.NewServer(thing)
	//server.DialWebSocket(user, passwd, "ws://127.0.0.1:8000/ws/1500", thing.Announce())
	server.DialWebSocket(user, passwd, "wss://demo.merliot.net/ws/1500", thing.Announce())
	//server.DialWebSocket(user, passwd, "wss://sw-poc.merliot.net/ws/1500", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
