package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/ps30m"
)

func main() {
	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	thing := ps30m.New("p1", "ps30m", "rpi-ps30m").(*ps30m.Ps30m)

	server := dean.NewServer(thing)
	server.BasicAuth(user, passwd)

	server.DialWebSocket(user, passwd, "wss://hub.merliot.net/ws/?ping-period=4", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
