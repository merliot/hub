package main

import (
	"github.com/merliot/hub/models/common"
	"github.com/merliot/hub/models/relays"
)

func main() {
	thing := relays.New("r1", "relays", "r1")
	common.Run(thing)
}
