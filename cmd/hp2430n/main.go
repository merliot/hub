package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/hp2430n"
)

func main() {
	port, _ := os.LookupEnv("PORT")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	thing := hp2430n.New("h1", "hp2430n", "rpi-hp2430n").(*hp2430n.Hp2430n)

	server := dean.NewServer(thing)
	server.BasicAuth(user, passwd)

	server.DialWebSocket(user, passwd, "ws://192.168.1.213:8000/ws/", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
