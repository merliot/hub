package main

import (
	"log"
	"os"

	"github.com/merliot/dean"
	poc "github.com/merliot/sw-poc"
	"github.com/merliot/sw-poc/modela"
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

	poc.Register("modela", modela.New)

	log.Fatal(server.ListenAndServe())
}
