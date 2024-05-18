package main

import (
	"log"
	"os"

	"github.com/merliot/hub"
)

//go:generate go run ./

func main() {
	hub := hub.NewHub("proto", "hub", "proto", "", "", "", "").(*hub.Hub)
	if err := hub.GenerateUf2s("../../uf2s/"); err != nil {
		log.Println("Error generating UF2s:", err)
		os.Exit(1)
	}
}
