// Merliot Hub
//
// go run ./cmd

package main

import (
	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

func main() {
	device.Models = models.AllModels
	device.Run()
}
