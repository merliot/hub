package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/ps30m"
)

func main() {
	thing := ps30m.New("230042", "ps30m", "poc1")

	server := dean.NewServer(thing)

	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	server.DialWebSocket(user, passwd, "ws://192.168.1.213:8000/ws/1500", thing.Announce())
	//server.DialWebSocket("user", "passwd", "wss://sw-poc.merliot.net/ws/1500", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
