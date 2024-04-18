package main

import (
	"log"
	"os"

	"github.com/merliot/hub"
)

//go:generate go run ../gen-models -input ../../models.json -output ./models.go
//go:generate gofmt -w ./models.go
//go:generate go run ./

func main() {
	hub := hub.New("proto", "hub", "proto").(*hub.Hub)

	for model, maker := range models {
		hub.RegisterModel(model, maker)
	}

	if err := hub.GenerateUf2s("../.."); err != nil {
		log.Println("Error generating UF2s:", err)
		os.Exit(1)
	}
}
