//go:generate sh -x -c "go run ./../gen-models/ ../../models.json ../../pkg/models/models.go"

//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ../../bin/mcp-server-linux-amd64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOARM=5 go build -ldflags '-s -w' -o ../../bin/mcp-server-linux-arm64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o ../../bin/mcp-server-darwin-amd64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o ../../bin/mcp-server-darwin-arm64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o ../../bin/mcp-server-windows-amd64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags '-s -w' -o ../../bin/mcp-server-windows-arm64 ./"

package main

import (
	"fmt"
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

	if err := MCPServer.ServeStdio(); err != nil {
		fmt.Println("MCP server error:", err)
		os.Exit(1)
	}
}
