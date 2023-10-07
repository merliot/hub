package gps

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type Gps struct {
	*common.Common
	Lat  float64
	Long float64
	targetStruct
}

type Update struct {
	Path string
	Lat  float64
	Long float64
}

var targets = []string{"demo", "x86-64", "rpi", "nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	g := &Gps{}
	g.Common = common.New(id, model, name, targets).(*common.Common)
	g.targetNew()
	return g
}

func (g *Gps) save(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
}

func (g *Gps) getState(msg *dean.Msg) {
	g.Path = "state"
	msg.Marshal(g).Reply()
}

func (g *Gps) update(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.save,
		"get/state": g.getState,
		"update":    g.update,
	}
}

func (g *Gps) Run(i *dean.Injector) {
	g.run(i)
}
