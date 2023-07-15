package main

import (
	"os"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/hub"
	"github.com/merliot/sw-poc/models/relays"
	"github.com/merliot/sw-poc/models/gps"
	"github.com/merliot/sw-poc/models/sense"
	"github.com/merliot/sw-poc/models/led"
)

func main() {

	hub := hub.New("swpoc01", "swpoc", "swpoc01")

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

	server.RegisterModel("relays", relays.New)
	server.RegisterModel("gps", gps.New)
	server.RegisterModel("sense", sense.New)
	server.RegisterModel("led", led.New)

	go server.ListenAndServe()
	server.Run()
}
