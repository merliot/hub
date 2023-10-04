package sign

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type Sign struct {
	*common.Common
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

func (s *Sign) save(msg *dean.Msg) {
	msg.Unmarshal(s).Broadcast()
}

func (s *Sign) getState(msg *dean.Msg) {
	s.Path = "state"
	msg.Marshal(s).Reply()
}

func (s *Sign) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     s.save,
		"get/state": s.getState,
	}
}

func (s *Sign) Run(i *dean.Injector) {
	s.run(i)
}
