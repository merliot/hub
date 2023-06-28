package main

import (
	"log"
	"os"

	"github.com/merliot/dean"
	poc "github.com/merliot/sw-poc"
	"github.com/merliot/sw-poc/models/relays"
	"github.com/merliot/sw-poc/models/gps"
)

func main() {

	poc := poc.New("swpoc01", "swpoc", "swpoc01").(*poc.Poc)

	server := dean.NewServer(poc)

	server.Addr = ":8000"
	if port, ok := os.LookupEnv("PORT"); ok {
		server.Addr = ":" + port
	}

	if user, ok := os.LookupEnv("USER"); ok {
		if passwd, ok := os.LookupEnv("PASSWD"); ok {
			server.BasicAuth(user, passwd)
		}
	}

	poc.Register("relays", relays.New)
	poc.Register("gps", gps.New)

	log.Fatal(server.ListenAndServe())
}
