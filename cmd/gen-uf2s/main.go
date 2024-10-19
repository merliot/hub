package main

import (
	"github.com/merliot/thing2"
	"github.com/merliot/thing2/models"
)

//go:generate go run ../gen-models/
//go:generate go run ./

func main() {
	thing2.Models = models.AllModels
	if err := thing2.GenerateUf2s("../../uf2s/"); err != nil {
		panic(err)
	}
}
