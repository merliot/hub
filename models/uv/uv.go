package uv

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type RiskLevel uint8

const (
	UVI_RISK_LOW RiskLevel = iota
	UVI_RISK_MODERATE
	UVI_RISK_HIGH
	UVI_RISK_VERY_HIGH
	UVI_RISK_EXTREME
)

type UV struct {
	*common.Common
	// UVA light intensity (irradiance) in milli Watt per square
	// meter (mW/(m*m))
	Intensity uint32
	RiskLevel
}

type Update struct {
	Path string
	Intensity uint32
	RiskLevel
}

var targets = []string{"nano-rp2040"}

func New(id, model, name string) dean.Thinger {
	println("NEW UV")
	u := &UV{}
	u.Common = common.New(id, model, name, targets).(*common.Common)
	return u
}

func (u *UV) save(msg *dean.Msg) {
	msg.Unmarshal(u).Broadcast()
}

func (u *UV) getState(msg *dean.Msg) {
	u.Path = "state"
	msg.Marshal(u).Reply()
}

func (u *UV) update(msg *dean.Msg) {
	msg.Unmarshal(u).Broadcast()
}

func (u *UV) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     u.save,
		"get/state": u.getState,
		"update":    u.update,
	}
}
