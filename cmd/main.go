package main

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/models"
)

func main() {
	hub.Models = models.AllModels
	hub.Run()
}
