package hub

import (
	"github.com/merliot/garage"
	"github.com/merliot/ps30m"
	"github.com/merliot/relays"
	"github.com/merliot/skeleton"
)

func (h *Hub) RegisterModels() {
	h.RegisterModel("garage", garage.New)
	h.RegisterModel("ps30m", ps30m.New)
	h.RegisterModel("relays", relays.New)
	h.RegisterModel("skeleton", skeleton.New)
}
