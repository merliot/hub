package sign

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type Display struct {
	Width  int16
	Height int16
}

type Terminal struct {
	Width  int16
	Height int16
}

type Sign struct {
	*common.Common
	Display Display
	Terminal Terminal
	Banner string
	targetStruct
}

var targets = []string{"wioterminal"}

func New(id, model, name string) dean.Thinger {
	println("NEW SIGN")
	s := &Sign{}
	s.Common = common.New(id, model, name, targets).(*common.Common)
	s.targetNew()
	return s
}

func (s *Sign) saveState(msg *dean.Msg) {
	msg.Unmarshal(s).Broadcast()
}

func (s *Sign) getState(msg *dean.Msg) {
	s.Path = "state"
	msg.Marshal(s).Reply()
}

func (s *Sign) save(msg *dean.Msg) {
	msg.Unmarshal(s)
	if s.IsMetal() {
		s.refresh()
		s.store()
	}
	msg.Broadcast()
}

func (s *Sign) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     s.saveState,
		"get/state": s.getState,
		"save":      s.save,
	}
}

func (s *Sign) Run(i *dean.Injector) {
	s.run(i)
}
