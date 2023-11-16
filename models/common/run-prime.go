//go:build prime

package common

import (
	"github.com/merliot/dean"
	"github.com/merliot/models/prime"
)

func Run(thing dean.Thinger) {
	prime := prime.New(thing)
	server := dean.NewServer(prime)
	server.Run()
}
