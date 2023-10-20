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

	thing := ps30m.New("230042", "ps30m", "poc1").(*ps30m.Ps30m)

	server := dean.NewServer(thing)
	server.BasicAuth(user, passwd)

	server.DialWebSocket(user, passwd, "ws://127.0.0.1:8000/ws/", thing.Announce())

	if port != "" {
		server.Addr = ":" + port
		go server.ListenAndServe()
	}

	server.Run()
}
