// Merliot Hub
//
// To run standalone:
//   $ go run ./cmd
//
// To generate x84-64 and rpi binaries, run:
//   $ go generate ./cmd

//go:generate sh -x -c "go run ./gen-models/ ../models.json ../pkg/models/models.go"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ../bin/device-x86-64 -tags x86_64 ./"
//go:generate sh -x -c "CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOARM=5 go build -ldflags '-s -w' -o ../bin/device-rpi -tags rpi ./"

package main

import (
	"github.com/merliot/hub/pkg/device"
	"github.com/merliot/hub/pkg/models"
)

func main() {
	port := devices.Getenv("PORT", "8000")
	addr := port + ":"
	server := device.NewServer(addr, models.AllModels)
	server.Run()
}
