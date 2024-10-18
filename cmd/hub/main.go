package main

import (
	"github.com/merliot/hub"
	"github.com/merliot/hub/models"
)

var device = `{
	"hub": {
		"Id": "hub,
		"Model": "hub",
		"Name": "Hub",
		"Children": [],
		"DeployParams": "target=x86-64&port=8000"
	}
}`

func main() {
	hub.Setenv("DEVICES", hub.Getenv("DEVICES", device))
	hub.Models = models.AllModels
	hub.Run()
}
