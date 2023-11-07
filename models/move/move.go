package move

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type Move  struct {
	*common.Common
	Ax, Ay, Az int32
	Gx, Gy, Gz int32
	targetStruct
}

var CALIBRATE_BEGIN = dean.ThingMsg{Path: "calibrate/begin"}
var CALIBRATE_END   = dean.ThingMsg{Path: "calibrate/end"}

type Update struct {
	Path       string
	Ax, Ay, Az int32
	Gx, Gy, Gz int32
}

var targets = []string{"nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW MOVE")
	m := &Move{}
	m.Common = common.New(id, model, name, targets).(*common.Common)
	m.targetNew()
	return m
}

func (m *Move) save(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
}

func (m *Move) getState(msg *dean.Msg) {
	m.Path = "state"
	msg.Marshal(m).Reply()
}

func (m *Move) update(msg *dean.Msg) {
	msg.Unmarshal(m).Broadcast()
}

func (m *Move) broadcast(msg *dean.Msg) {
	msg.Broadcast()
}

func (m *Move) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":           m.save,
		"get/state":       m.getState,
		"update":          m.update,
		"calibrate/begin": m.broadcast,
		"calibrate/end":   m.broadcast,
	}
}
