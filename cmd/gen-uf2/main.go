package main

import (
	"log"
	"os"

	"github.com/merliot/hub"
)

func main() {
	hub := hub.New("proto", "hub", "proto").(*hub.Hub)
	if err := hub.GenerateUf2s(); err != nil {
		log.Println("Error generating UF2s:", err)
		os.Exit(1)
	}
}
