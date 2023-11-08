package charge

import (
	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

type Charge struct {
	*common.Common
	targetStruct
}

func New(id, model, name string, targets []string) dean.Thinger {
	println("NEW CHARGE")
	c := &Charge{}
	c.Common = common.New(id, model, name, targets).(*common.Common)
	c.targetNewX()
	return c
}
