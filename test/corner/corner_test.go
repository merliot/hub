package main

import (
	"os"
	"testing"
	"time"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
	"github.com/stretchr/testify/require"
)

var hub = `{
	"gadget1": {
		"Id": "gadget1",
		"Model": "gadget",
		"Name": "Gadget",
		"Children": null,
		"DeployParams": "target=x86-64&port=&Bottles=99"
	},
	"subhub1": {
		"Id": "subhub1",
		"Model": "hub",
		"Name": "Hub",
		"Children": null,
		"DeployParams": "target=x86-64\u0026port=8000"
	},
	"hub": {
		"Id": "hub",
		"Model": "hub",
		"Name": "Hub",
		"Children": [
			"gadget1",
			"subhub1"
		],
		"DeployParams": "target=x86-64\u0026port=8000"
	}
}`

func TestOddOptions(t *testing.T) {
	err := os.WriteFile("devices.json", []byte(hub), 0644)
	require.NoError(t, err)

	// Run a hub
	hubby := device.NewServer(
		device.WithPort(10),
		device.WithPingPeriod(1),
		device.WithAutoSave("true"),
		device.WithModels(models.AllModels),
		device.WithKeepBuilds("true"),
		device.WithDevicesFile("devices.json"),
	)
	go hubby.Run()
	time.Sleep(time.Second)

	err = os.RemoveAll("devices.json")
	require.NoError(t, err)
}
