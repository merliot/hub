//go:build tinygo

package common

import (
	"github.com/meriot/dean"
)

var (
	ssid string
	pass string
)

func Run(thing dean.Thinger) {
	tinynet.NetConnect(ssid, pass)
	prime := prime.New(thing)
	runner := dean.NewServer(prime)
	runner.DialWebSocket()
	runner.Run()
}
