package main

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/models"
)

//go:generate go run ../gen-models/
//go:generate go run ./

func main() {
	hub.Models = models.AllModels
	if err := hub.GenerateUf2s("../../uf2s/"); err != nil {
		panic(err)
	}
}
