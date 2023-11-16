//go:build !tinygo && !prime

package common

import (
	"github.com/merliot/dean"
)

func Run(thing dean.Thinger) {
	server := dean.NewServer(thing)
	server.DialWebSocket()
	server.Run()
}
