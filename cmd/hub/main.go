package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/hub"
	"github.com/merliot/sw-poc/models/relays"
	"github.com/merliot/sw-poc/models/gps"
	"github.com/merliot/sw-poc/models/sense"
	"github.com/merliot/sw-poc/models/led"
	"github.com/merliot/sw-poc/models/ps30m"
)

func main() {

	hub := hub.New("swpoc01", "swpoc", "swpoc01").(*hub.Hub)

	server := dean.NewServer(hub)

	server.Addr = ":8000"
	if port, ok := os.LookupEnv("PORT"); ok {
		server.Addr = ":" + port
	}

	if user, ok := os.LookupEnv("USER"); ok {
		if passwd, ok := os.LookupEnv("PASSWD"); ok {
			server.BasicAuth(user, passwd)
		}
	}

	hub.Register("relays", relays.New)
	hub.Register("gps", gps.New)
	hub.Register("sense", sense.New)
	hub.Register("led", led.New)
	hub.Register("ps30m", ps30m.New)

	go server.ListenAndServe()
	server.Run()
}
