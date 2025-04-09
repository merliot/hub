package main

import (
	"os"

	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

func main() {

	url, _ := os.LookupEnv("HUB_URL")
	user, _ := os.LookupEnv("USER")
	passwd, _ := os.LookupEnv("PASSWD")

	MCPServer := device.NewMCPServer(
		device.WithMCPModels(models.AllModels),
		device.WithMCPHubURL(url),
		device.WithMCPUser(user),
		device.WithMCPPasswd(passwd),
	)

	MCPServer.ServeStdio()
}
