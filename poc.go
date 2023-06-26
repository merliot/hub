package poc

import (
	"github.com/merliot/dean"
	"github.com/merliot/dean-lib/hub"
)

type Poc struct {
	*hub.Hub
}

func New(id, model, name string) dean.Thinger {
	println("NEW POC")
	return &Poc{
		Hub: hub.New(id, model, name).(*hub.Hub),
	}
}
