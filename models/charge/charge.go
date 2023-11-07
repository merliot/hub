package charge

import (
	"embed"
	"html/template"

	"github.com/merliot/dean"
	"github.com/merliot/hub/models/common"
)

//go:embed *
var fs embed.FS

type Charge struct {
	*common.Common
	templates *template.Template
}

func New(id, model, name string, targets []string) dean.Thinger {
	println("NEW CHARGE")
	c := &Charge{}
	c.Common = common.New(id, model, name, targets).(*common.Common)
	c.CompositeFs.AddFS(fs)
	c.templates = c.CompositeFs.ParseFS("template/*")
	return c
}
