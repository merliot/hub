package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/gps"
)

func main() {
	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")
	demo, _ := os.LookupEnv("DEMO")

	thing := gps.New("foo", "gps", "foo").(*gps.Gps)

	if demo != "" {
		thing.Demo()
	}

	server := dean.NewServer(thing)
	server.BasicAuth(user, passwd)

	server.DialWebSocket(user, passwd, "ws://127.0.0.1:8000/ws/1500", thing.Announce())
	//server.DialWebSocket(user, passwd, "wss://demo.merliot.net/ws/1500", thing.Announce())
	//server.DialWebSocket(user, passwd, "wss://hub.merliot.net/ws/1500", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
